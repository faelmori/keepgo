package linux

import (
	"bytes"
	"errors"
	"fmt"
	. "github.com/faelmori/keepgo/internal"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"text/template"
	"time"
)

func isOpenRC() bool {
	if _, err := exec.LookPath("openrc-init"); err == nil {
		return true
	}
	if _, err := os.Stat("/etc/inittab"); err == nil {
		filerc, err := os.Open("/etc/inittab")
		if err != nil {
			return false
		}
		defer filerc.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(filerc)
		contents := buf.String()

		re := regexp.MustCompile(`::sysinit:.*openrc.*sysinit`)
		matches := re.FindStringSubmatch(contents)
		if len(matches) > 0 {
			return true
		}
		return false
	}
	return false
}

type Openrc struct {
	i        Interface
	platform string
	*Config
}

func (s *Openrc) Logger(errs chan<- error) (Logger, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Openrc) String() string {
	if len(s.DisplayName) > 0 {
		return s.DisplayName
	}
	return s.Name
}
func (s *Openrc) Platform() string {
	return s.platform
}
func (s *Openrc) template() *template.Template {
	customScript := s.Option.String(OptionOpenRCScript, "")

	if customScript != "" {
		return template.Must(template.New("").Funcs(tf).Parse(customScript))
	}
	return template.Must(template.New("").Funcs(tf).Parse(OpenRCScript))
}

func newOpenRCService(i Interface, platform string, c *Config) (Service, error) {
	s := &Openrc{
		i:        i,
		platform: platform,
		Config:   c,
	}
	return s, nil
}

var errNoUserServiceOpenRC = errors.New("user services are not supported on OpenRC")

func (s *Openrc) configPath() (cp string, err error) {
	if s.Option.Bool(OptionUserService, OptionUserServiceDefault) {
		err = errNoUserServiceOpenRC
		return
	}
	cp = "/etc/init.d/" + s.Config.Name
	return
}
func (s *Openrc) Install() error {
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

	err = os.Chmod(confPath, 0755)
	if err != nil {
		return err
	}

	path, err := s.ExecPath()
	if err != nil {
		return err
	}

	var to = &struct {
		*Config
		Path         string
		LogDirectory string
	}{
		s.Config,
		path,
		s.Option.String(OptionLogDirectory, DefaultLogDirectory),
	}

	err = s.template().Execute(f, to)
	if err != nil {
		return err
	}
	// run rc-update
	return s.runAction("add")
}
func (s *Openrc) Uninstall() error {
	confPath, err := s.configPath()
	if err != nil {
		return err
	}
	if err := os.Remove(confPath); err != nil {
		return err
	}
	return s.runAction("delete")
}
func (s *Openrc) GetLogger(errs chan<- error) (Logger, error) {
	if Interactive() {
		return nil, nil
	}
	return s.SystemLogger(errs)
}
func (s *Openrc) SystemLogger(errs chan<- error) (Logger, error) {
	return newSysLogger(s.Name, errs)
}
func (s *Openrc) Run() (err error) {
	err = s.i.Start(s)
	if err != nil {
		return err
	}
	s.(OptionRunWait, func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		err = s.Stop()
	})
	return s.i.Stop(s)
}
func (s *Openrc) Status() (Status, error) {
	// rc-service uses the errno library for its exit codes:
	// errno 0 = service started
	// errno 1 = EPERM 1 Operation not permitted
	// errno 2 = ENOENT 2 No such file or directory
	// errno 3 = ESRCH 3 No such process
	// for more info, see https://man7.org/linux/man-pages/man3/errno.3.html
	_, out, err := runWithOutput("rc-service", s.Name, "status")
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			exitCode := exiterr.ExitCode()
			switch {
			case exitCode == 1:
				return StatusUnknown, err
			case exitCode == 2:
				return StatusUnknown, ErrNotInstalled
			case exitCode == 3:
				return StatusStopped, nil
			default:
				return StatusUnknown, fmt.Errorf("unknown error: %v - %v", out, err)
			}
		} else {
			return StatusUnknown, err
		}
	}
	return StatusRunning, nil
}
func (s *Openrc) Start() error {
	return run("rc-service", s.Name, "start")
}
func (s *Openrc) Stop() error {
	return run("rc-service", s.Name, "stop")
}
func (s *Openrc) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}
func (s *Openrc) runAction(action string) error {
	return s.run(action, s.Name)
}
func (s *Openrc) run(action string, args ...string) error {
	return run("rc-update", append([]string{action}, args...)...)
}

const OpenRCScript = `#!/sbin/openrc-run
supervisor=supervise-daemon
name="{{.DisplayName}}"
description="{{.Description}}"
command={{.Path|cmdEscape}}
{{- if .Arguments }}
command_args="{{range .Arguments}}{{.}} {{end}}"
{{- end }}
name=$(basename $(readlink -f $command))
supervise_daemon_args="--stdout {{.LogDirectory}}/${name}.log --stderr {{.LogDirectory}}/${name}.err"

{{range $k, $v := .EnvVars -}}
export {{$k}}={{$v}}
{{end -}}

{{- if .Dependencies }}
depend() {
{{- range $i, $dep := .Dependencies}} 
{{"\t"}}{{$dep}}{{end}}
}
{{- end}}
`
