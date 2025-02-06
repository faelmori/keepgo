// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package linux

import (
	"bytes"
	"errors"
	"fmt"
	. "github.com/faelmori/keepgo/internal"
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

func isSystemd() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false
	}
	if _, err := os.Stat("/proc/1/comm"); err == nil {
		filerc, err := os.Open("/proc/1/comm")
		if err != nil {
			return false
		}
		defer filerc.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(filerc)
		contents := buf.String()

		if strings.Trim(contents, " \r\n") == "systemd" {
			return true
		}
	}
	return false
}

type systemd struct {
	i        Controller
	platform string
	*Config
}

func newSystemdService(i Controller, platform string, c *Config) (Service, error) {
	s := &systemd{
		i:        i,
		platform: platform,
		Config:   c,
	}

	return s, nil
}

func (s *systemd) String() string {
	if len(s.DisplayName) > 0 {
		return s.DisplayName
	}
	return s.Name
}
func (s *systemd) Platform() string {
	return s.platform
}
func (s *systemd) ConfigPath() (cp string, err error) {
	if !s.IsUserService() {
		cp = "/etc/systemd/system/" + s.UnitName()
		return
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	systemdUserDir := filepath.Join(homeDir, ".config/systemd/user")
	err = os.MkdirAll(systemdUserDir, os.ModePerm)
	if err != nil {
		return
	}
	cp = filepath.Join(systemdUserDir, s.UnitName())
	return
}
func (s *systemd) UnitName() string {
	return s.Config.Name + ".service"
}
func (s *systemd) GetSystemdVersion() int64 {
	_, out, err := s.runWithOutput("systemctl", "--Version")
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
func (s *systemd) HasOutputFileSupport() bool {
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
func (s *systemd) GetTemplate() *template.Template {
	customScript := s.Option.String(OptionSystemdScript, "")

	if customScript != "" {
		return template.Must(template.New("").Funcs(tf).Parse(customScript))
	}
	return template.Must(template.New("").Funcs(tf).Parse(systemdScript))
}
func (s *systemd) IsUserService() bool {
	return s.Option.Bool(OptionUserService, OptionUserServiceDefault)
}
func (s *systemd) Install() error {
	confPath, err := s.ConfigPath()
	if err != nil {
		return err
	}
	_, err = os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("Init already exists: %s", confPath)
	}

	f, err := os.OpenFile(confPath, os.O_WRONLY|os.O_CREATE, 0644)
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
		s.Option.String(OptionReloadSignal, ""),
		s.Option.String(OptionPIDFile, ""),
		s.Option.Int(OptionLimitNOFILE, OptionLimitNOFILEDefault),
		s.Option.String(OptionRestart, "always"),
		s.Option.String(OptionSuccessExitStatus, ""),
		s.Option.Bool(OptionLogOutput, OptionLogOutputDefault),
		s.Option.String(OptionLogDirectory, DefaultLogDirectory),
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
func (s *systemd) Uninstall() error {
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
func (s *systemd) GetLogger(errs chan<- error) (Logger, error) {
	if Interactive() {
		return newSysLogger(s.Name, errs)
	}
	return s.SystemLogger(errs)
}
func (s *systemd) SystemLogger(errs chan<- error) (Logger, error) {
	return newSysLogger(s.Name, errs)
}
func (s *systemd) Run() (err error) {
	err = s.Start()
	if err != nil {
		return err
	}

	s.Option.FuncSingle(OptionRunWait, func() {
		var sigChan = make(chan os.Signal, 3)
		signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
		<-sigChan
	})()

	return s.Stop()
}
func (s *systemd) Status() (Status, error) {
	exitCode, out, err := s.runWithOutput("systemctl", "is-active", s.UnitName())
	if exitCode == 0 && err != nil {
		return StatusUnknown, err
	}

	switch {
	case strings.HasPrefix(out, "active"):
		return StatusRunning, nil
	case strings.HasPrefix(out, "inactive"):
		// inactive can also mean its not installed, check unit files
		exitCode, out, err := s.runWithOutput("systemctl", "list-unit-files", "-t", "service", s.UnitName())
		if exitCode == 0 && err != nil {
			return StatusUnknown, err
		}
		if strings.Contains(out, s.Name) {
			// unit file exists, installed but not running
			return StatusStopped, nil
		}
		// no unit file
		return StatusUnknown, ErrNotInstalled
	case strings.HasPrefix(out, "activating"):
		return StatusRunning, nil
	case strings.HasPrefix(out, "failed"):
		return StatusUnknown, errors.New("service in failed state")
	default:
		return StatusUnknown, ErrNotInstalled
	}
}
func (s *systemd) Start() error {
	return s.runAction("start")
}
func (s *systemd) Stop() error {
	return s.runAction("stop")
}
func (s *systemd) Restart() error {
	return s.runAction("restart")
}
func (s *systemd) runWithOutput(command string, arguments ...string) (int, string, error) {
	if s.IsUserService() {
		arguments = append(arguments, "--user")
	}
	return runWithOutput(command, arguments...)
}
func (s *systemd) run(action string, args ...string) error {
	if s.IsUserService() {
		return run("systemctl", append([]string{action, "--user"}, args...)...)
	}
	return run("systemctl", append([]string{action}, args...)...)
}
func (s *systemd) runAction(action string) error {
	return s.run(action, s.UnitName())
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
