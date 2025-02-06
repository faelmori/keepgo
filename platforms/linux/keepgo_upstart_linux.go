// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package linux

import (
	"errors"
	"fmt"
	internal "github.com/faelmori/keepgo/internal"
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
	i        internal.Controller
	platform string
	*internal.Config
}

func newUpstartService(i internal.Controller, platform string, c *internal.Config) (internal.Service, error) {
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

func (s *Upstart) configPath() (cp string, err error) {
	if s.Option.Bool(internal.OptionUserService, internal.OptionUserServiceDefault) {
		err = errNoUserServiceUpstart
		return
	}
	cp = "/etc/init/" + s.Config.Name + ".conf"
	return
}
func (s *Upstart) hasKillStanza() bool {
	defaultValue := true
	version := s.getUpstartVersion()
	if version == nil {
		return defaultValue
	}

	maxVersion := []int{0, 6, 5}
	if matches, err := VersionAtMost(version, maxVersion); err != nil || matches {
		return false
	}

	return defaultValue
}
func (s *Upstart) hasSetUIDStanza() bool {
	defaultValue := true
	version := s.getUpstartVersion()
	if version == nil {
		return defaultValue
	}

	maxVersion := []int{1, 4, 0}
	if matches, err := VersionAtMost(version, maxVersion); err != nil || matches {
		return false
	}

	return defaultValue
}
func (s *Upstart) getUpstartVersion() []int {
	_, out, err := runWithOutput("/sbin/initctl", "--Version")
	if err != nil {
		return nil
	}

	re := regexp.MustCompile(`initctl \(Upstart (\d+.\d+.\d+)\)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) != 2 {
		return nil
	}

	return internal.ParseVersion(matches[1])
}
func (s *Upstart) template() *template.Template {
	customScript := s.Option.string(optionUpstartScript, "")

	if customScript != "" {
		return template.Must(template.New("").Funcs(tf).Parse(customScript))
	} else {
		return template.Must(template.New("").Funcs(tf).Parse(upstartScript))
	}
}
func (s *Upstart) Install() error {
	confPath, err := s.configPath()
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
		*Config
		Path            string
		HasKillStanza   bool
		HasSetUIDStanza bool
		LogOutput       bool
		LogDirectory    string
	}{
		s.Config,
		path,
		s.hasKillStanza(),
		s.hasSetUIDStanza(),
		s.Option.Bool(OptionLogOutput, OptionLogOutputDefault),
		s.Option.String(OptionLogDirectory, DefaultLogDirectory),
	}

	return s.template().Execute(f, to)
}
func (s *Upstart) Uninstall() error {
	cp, err := s.configPath()
	if err != nil {
		return err
	}
	if err := os.Remove(cp); err != nil {
		return err
	}
	return nil
}
func (s *Upstart) Logger(errs chan<- error) (Logger, error) {
	if Interactive() {
		return nil, nil
	}
	return s.SystemLogger(errs)
}
func (s *Upstart) SystemLogger(errs chan<- error) (Logger, error) {
	return newSysLogger(s.Name, errs)
}
func (s *Upstart) Run() (err error) {
	err = s.i.Start(s)
	if err != nil {
		return err
	}

	s.Option.FuncSingle(OptionRunWait, func() {
		var sigChan = make(chan os.Signal, 3)
		signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
		<-sigChan
	})()

	return s.i.Stop(s)
}
func (s *Upstart) Status() (Status, error) {
	exitCode, out, err := runWithOutput("initctl", "status", s.Name)
	if exitCode == 0 && err != nil {
		return StatusUnknown, err
	}

	switch {
	case strings.HasPrefix(out, fmt.Sprintf("%s start/running", s.Name)):
		return StatusRunning, nil
	case strings.HasPrefix(out, fmt.Sprintf("%s stop/waiting", s.Name)):
		return StatusStopped, nil
	default:
		return StatusUnknown, ErrNotInstalled
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
