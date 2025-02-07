package linux

import (
	"github.com/faelmori/keepgo/service"
	"os/exec"
	"time"
)

func NewUnixService(i service.Controller, platform string, c *service.Config) (service.Service, error) {
	return &unixService{Name: c.Name}, nil
}

type unixService struct {
	Name string
}

func (s *unixService) Run() error                                          { panic("implement me") }
func (s *unixService) Install() error                                      { panic("implement me") }
func (s *unixService) Uninstall() error                                    { panic("implement me") }
func (s *unixService) GetLogger(errs chan<- error) (service.Logger, error) { panic("implement me") }
func (s *unixService) SystemLogger(errs chan<- error) (service.Logger, error) {
	panic("implement me")
}
func (s *unixService) String() string                  { return s.Name }
func (s *unixService) Platform() string                { return "unix" }
func (s *unixService) Status() (service.Status, error) { panic("implement me") }
func (s *unixService) Start() error                    { return runUnixCommand("/bin/systemctl", "start", s.Name) }
func (s *unixService) Stop() error                     { return runUnixCommand("/bin/systemctl", "stop", s.Name) }
func (s *unixService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}

func runUnixCommand(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	return cmd.Run()
}

func IsUnix() bool {
	_, err := exec.LookPath("systemctl")
	return err == nil
}
