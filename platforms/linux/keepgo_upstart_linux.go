// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package linux

import (
	"errors"
	"fmt"
	"github.com/faelmori/keepgo/internal"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"text/template"
)

func IsUpstart() bool {
	if _, err := os.Stat("/sbin/upstart-udev-bridge"); err == nil {
		return true
	}
	if _, err := os.Stat("/sbin/initctl"); err == nil {
		if _, out, err := runWithOutput("/sbin/initctl", "--Version"); err == nil {
			if strings.Contains(out, "initctl (upstart") {
				return true
			}
		}
	}
	return false
}

type Upstart struct {
	i        keepgo.Controller
	platform string
	*keepgo.Config
}

func newUpstartService(i keepgo.Controller, platform string, c *keepgo.Config) (keepgo.Service, error) {
	s := &Upstart{
		i:        i,
		platform: platform,
		Config:   c,
	}

	return s, nil
}
func (s *Upstart) String() string {
	if len(s.DisplayName) > 0 {
		return s.DisplayName
	}
	return s.Name
}
func (s *Upstart) Platform() string {
	return s.platform
}

// Upstart has some support for user services in graphical sessions.
// Due to the mix of actual support for user services over versions, just don't bother.
// Upstart will be replaced by systemd in most cases anyway.
var errNoUserServiceUpstart = errors.New("User services are not supported on Upstart.")

func (s *Upstart) ConfigPath() (cp string, err error) {
	if s.Option.Bool(keepgo.OptionUserService, keepgo.OptionUserServiceDefault) {
		err = errNoUserServiceUpstart
		return
	}
	cp = "/etc/init/" + s.Config.Name + ".conf"
	return
}
func (s *Upstart) HasKillStanza() bool {
	defaultValue := true
	version := s.GetUpstartVersion()
	if version == nil {
		return defaultValue
	}

	maxVersion := []int{0, 6, 5}
	if matches, err := keepgo.VersionAtMost(version, maxVersion); err != nil || matches {
		return false
	}

	return defaultValue
}
func (s *Upstart) HasSetUIDStanza() bool {
	defaultValue := true
	version := s.GetUpstartVersion()
	if version == nil {
		return defaultValue
	}

	maxVersion := []int{1, 4, 0}
	if matches, err := keepgo.VersionAtMost(version, maxVersion); err != nil || matches {
		return false
	}

	return defaultValue
}
func (s *Upstart) GetUpstartVersion() []int {
	_, out, err := runWithOutput("/sbin/initctl", "--Version")
	if err != nil {
		return nil
	}

	re := regexp.MustCompile(`initctl \(Upstart (\d+.\d+.\d+)\)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) != 2 {
		return nil
	}

	return keepgo.ParseVersion(matches[1])
}
func (s *Upstart) GetTemplate() *template.Template {
	customScript := s.Option.String(keepgo.OptionUpstartScript, "")

	if customScript != "" {
		return template.Must(template.New("").Funcs(tf).Parse(customScript))
	} else {
		return template.Must(template.New("").Funcs(tf).Parse(upstartScript))
	}
}
func (s *Upstart) Install() error {
	confPath, err := s.ConfigPath()
	if err != nil {
		return err
	}
	_, err = os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("Init already exists: %s", confPath)
	}

	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	path, err := s.ExecPath()
	if err != nil {
		return err
	}

	var to = &struct {
		*keepgo.Config
		Path            string
		HasKillStanza   bool
		HasSetUIDStanza bool
		LogOutput       bool
		LogDirectory    string
	}{
		s.Config,
		path,
		s.HasKillStanza(),
		s.HasSetUIDStanza(),
		s.Option.Bool(keepgo.OptionLogOutput, keepgo.OptionLogOutputDefault),
		s.Option.String(keepgo.OptionLogDirectory, DefaultLogDirectory),
	}

	return s.GetTemplate().Execute(f, to)
}
func (s *Upstart) Uninstall() error {
	cp, err := s.ConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(cp); err != nil {
		return err
	}
	return nil
}
func (s *Upstart) GetLogger(errs chan<- error) (keepgo.Logger, error) {
	if keepgo.Interactive() {
		return nil, nil
	}
	return s.SystemLogger(errs)
}
func (s *Upstart) SystemLogger(errs chan<- error) (keepgo.Logger, error) {
	return newSysLogger(s.Name, errs)
}
func (s *Upstart) Run() (err error) {
	err = s.i.Start(s)
	if err != nil {
		return err
	}

	s.Option.FuncSingle(keepgo.OptionRunWait, func() {
		var sigChan = make(chan os.Signal, 3)
		signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
		<-sigChan
	})()

	return s.i.Stop(s)
}
func (s *Upstart) Status() (keepgo.Status, error) {
	exitCode, out, err := runWithOutput("initctl", "status", s.Name)
	if exitCode == 0 && err != nil {
		return keepgo.StatusUnknown, err
	}

	switch {
	case strings.HasPrefix(out, fmt.Sprintf("%s start/running", s.Name)):
		return keepgo.StatusRunning, nil
	case strings.HasPrefix(out, fmt.Sprintf("%s stop/waiting", s.Name)):
		return keepgo.StatusStopped, nil
	default:
		return keepgo.StatusUnknown, keepgo.ErrNotInstalled
	}
}
func (s *Upstart) Start() error {
	return run("initctl", "start", s.Name)
}
func (s *Upstart) Stop() error {
	return run("initctl", "stop", s.Name)
}
func (s *Upstart) Restart() error {
	return run("initctl", "restart", s.Name)
}

// The Upstart script should stop with an INT or the Go runtime will terminate
// the program before the Stop handler can run.
const upstartScript = `# {{.Description}}

{{if .DisplayName}}description    "{{.DisplayName}}"{{end}}

{{if .HasKillStanza}}kill signal INT{{end}}
{{if .ChRoot}}chroot {{.ChRoot}}{{end}}
{{if .WorkingDirectory}}chdir {{.WorkingDirectory}}{{end}}
start on filesystem or runlevel [2345]
stop on runlevel [!2345]

{{if and .UserName .HasSetUIDStanza}}setuid {{.UserName}}{{end}}

respawn
respawn limit 10 5
umask 022

console none

pre-start script
    test -x {{.Path}} || { stop; exit 0; }
end script

# Start
script
	{{if .LogOutput}}
	stdout_log="{{.LogDirectory}}/{{.Name}}.out"
	stderr_log="{{.LogDirectory}}/{{.Name}}.err"
	{{end}}
	
	if [ -f "/etc/sysconfig/{{.Name}}" ]; then
		set -a
		source /etc/sysconfig/{{.Name}}
		set +a
	fi

	exec {{if and .UserName (not .HasSetUIDStanza)}}sudo -E -u {{.UserName}} {{end}}{{.Path}}{{range .Arguments}} {{.|cmd}}{{end}}{{if .LogOutput}} >> $stdout_log 2>> $stderr_log{{end}}
end script
`
