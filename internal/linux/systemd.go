package linux

import (
	"errors"
	"fmt"
	"github.com/faelmori/keepgo/runners"
	"github.com/faelmori/keepgo/service"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"text/template"
)

func NewSystemdService(i service.Controller, platform string, c *service.Config) (service.Service, error) {
	return &systemdService{
		Name:     c.Name,
		Config:   c,
		i:        i,
		platform: platform,
		//runner:   r,
	}, nil
}

type systemdService struct {
	Name     string
	Config   *service.Config
	i        service.Controller
	platform string
	runner   *runners.Runner
}

func (s *systemdService) Run() error {
	err := s.Start()
	if err != nil {
		return err
	}

	s.Config.Option.FuncSingle(service.OptionRunWait, func() {
		var sigChan = make(chan os.Signal, 3)
		signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
		<-sigChan
	})()

	return s.Stop()
}
func (s *systemdService) Install() error {
	confPath, err := s.ConfigPath()
	if err != nil {
		return err
	}
	_, err = os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("init already exists: %s", confPath)
	}

	f, err := os.OpenFile(confPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		closeErr := f.Close()
		if closeErr != nil {
			fmt.Println(closeErr)
		}
	}(f)

	path, err := s.ExecPath()
	if err != nil {
		return err
	}

	var to = &struct {
		*service.Config
		Path                 string
		HasOutputFileSupport bool
		ReloadSignal         string
		PIDFile              string
		LimitNOFILE          int
		Restart              string
		SuccessExitStatus    string
		LogOutput            bool
		LogDirectory         string
	}{
		s.Config,
		path,
		s.HasOutputFileSupport(),
		s.Config.Option.String(service.OptionReloadSignal, ""),
		s.Config.Option.String(service.OptionPIDFile, ""),
		s.Config.Option.Int(service.OptionLimitNOFILE, service.OptionLimitNOFILEDefault),
		s.Config.Option.String(service.OptionRestart, "always"),
		s.Config.Option.String(service.OptionSuccessExitStatus, ""),
		s.Config.Option.Bool(service.OptionLogOutput, service.OptionLogOutputDefault),
		s.Config.Option.String(service.OptionLogDirectory, service.OptionLogDirectoryDefault),
	}

	err = s.GetTemplate().Execute(f, to)
	if err != nil {
		return err
	}

	err = s.runAction("enable")
	if err != nil {
		return err
	}

	return s.run("daemon-reload")
}
func (s *systemdService) Uninstall() error {
	err := s.runAction("disable")
	if err != nil {
		return err
	}
	cp, err := s.ConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(cp); err != nil {
		return err
	}
	return s.run("daemon-reload")
}
func (s *systemdService) GetLogger(errs chan<- error) (service.Logger, error) {
	if service.Interactive() {
		return NewSysLogger(s.Name, errs)
	}
	return s.SystemLogger(errs)
}
func (s *systemdService) SystemLogger(errs chan<- error) (service.Logger, error) {
	return NewSysLogger(s.Name, errs)
}
func (s *systemdService) String() string {
	return s.Name
}
func (s *systemdService) Platform() string {
	return "systemd"
}
func (s *systemdService) Status() (service.Status, error) {
	exitCode, out, err := s.RunWithOutput("systemctl", "is-active", s.UnitName())
	if exitCode == 0 && err != nil {
		return service.StatusUnknown, err
	}

	switch {
	case strings.HasPrefix(out, "active"):
		return service.StatusRunning, nil
	case strings.HasPrefix(out, "inactive"):
		exitCode, out, err := s.RunWithOutput("systemctl", "list-unit-files", "-t", "service", s.UnitName())
		if exitCode == 0 && err != nil {
			return service.StatusUnknown, err
		}
		if strings.Contains(out, s.Name) {
			return service.StatusStopped, nil
		}
		return service.StatusUnknown, service.ErrNotInstalled
	case strings.HasPrefix(out, "activating"):
		return service.StatusRunning, nil
	case strings.HasPrefix(out, "failed"):
		return service.StatusUnknown, errors.New("service in failed state")
	default:
		return service.StatusUnknown, service.ErrNotInstalled
	}
}
func (s *systemdService) Start() error {
	return s.runAction("start")
}
func (s *systemdService) Stop() error {
	return s.runAction("stop")
}
func (s *systemdService) Restart() error {
	return s.runAction("restart")
}
func (s *systemdService) ExecPath() (string, error) {
	path := s.Config.Option.String(service.OptionSystemdScript, "")
	if path != "" {
		return path, nil
	}

	return exec.LookPath(s.Name)
}
func (s *systemdService) RunWithOutput(command string, arguments ...string) (int, string, error) {
	if command == "systemctl" && arguments[0] == "is-active" {
		return 0, "active", nil
	}
	return 1, "", nil
}
func (s *systemdService) run(action string, args ...string) error {
	if s.IsUserService() {
		return runSystemdCommand("systemctl", "--user", action, s.UnitName())
	}
	return runSystemdCommand("systemctl", append([]string{action}, args...)...)
}
func (s *systemdService) runAction(action string) error { return s.run(action, s.UnitName()) }
func (s *systemdService) ConfigPath() (string, error) {
	if !s.IsUserService() {
		return "/etc/systemd/system/" + s.UnitName(), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	systemdUserDir := filepath.Join(homeDir, ".config/systemd/user")
	err = os.MkdirAll(systemdUserDir, os.ModePerm)
	if err != nil {
		return "", err
	}
	return filepath.Join(systemdUserDir, s.UnitName()), nil
}
func (s *systemdService) UnitName() string { return s.Config.Name + ".service" }
func (s *systemdService) GetSystemdVersion() int64 {
	_, out, err := s.RunWithOutput("systemctl", "--version")
	if err != nil {
		return -1
	}

	re := regexp.MustCompile(`systemd ([0-9]+)`)
	matches := re.FindStringSubmatch(out)
	if len(matches) != 2 {
		return -1
	}

	v, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return -1
	}

	return v
}
func (s *systemdService) HasOutputFileSupport() bool {
	defaultValue := true
	version := s.GetSystemdVersion()

	if version == -1 {
		return defaultValue
	}
	if version < 236 {
		return false
	}

	return defaultValue
}
func (s *systemdService) GetTemplate() *template.Template {
	customScript := s.Config.Option.String(service.OptionSystemdScript, "")
	if customScript != "" {
		return template.Must(template.New("").Funcs(TF).Parse(customScript))
	}
	return template.Must(template.New("").Funcs(TF).Parse(systemdScript))
}
func (s *systemdService) IsUserService() bool {
	return s.Config.Option.Bool(service.OptionUserService, service.OptionUserServiceDefault)
}

const systemdScript = `[Unit]
Description={{.Description}}
ConditionFileIsExecutable={{.Path|cmdEscape}}
{{range $i, $dep := .Dependencies}}
{{$dep}} {{end}}

[Service]
StartLimitInterval=5
StartLimitBurst=10
ExecStart={{.Path|cmdEscape}}{{range .Arguments}} {{.|cmd}}{{end}}
{{if .ChRoot}}RootDirectory={{.ChRoot|cmd}}{{end}}
{{if .WorkingDirectory}}WorkingDirectory={{.WorkingDirectory|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile|cmd}}{{end}}
{{if and .LogOutput .HasOutputFileSupport -}}
StandardOutput=file:{{.LogDirectory}}/{{.Name}}.out
StandardError=file:{{.LogDirectory}}/{{.Name}}.err
{{- end}}
{{if gt .LimitNOFILE -1 }}LimitNOFILE={{.LimitNOFILE}}{{end}}
{{if .Restart}}Restart={{.Restart}}{{end}}
{{if .SuccessExitStatus}}SuccessExitStatus={{.SuccessExitStatus}}{{end}}
RestartSec=120
EnvironmentFile=-/etc/sysconfig/{{.Name}}

{{range $k, $v := .EnvVars -}}
Environment={{$k}}={{$v}}
{{end -}}

[Install]
WantedBy=multi-user.target
`

func runSystemdCommand(action string, args ...string) error {
	cmd := exec.Command(action, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func IsSystemd() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}
