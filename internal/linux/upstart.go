package linux

import (
	"github.com/faelmori/keepgo/service"
	"os/exec"
	"time"
)

func NewUpstartService(i service.Controller, platform string, c *service.Config) (service.Service, error) {
	return &upstartService{Name: c.Name}, nil
}

type upstartService struct {
	Name string
}

func (s *upstartService) Run() error                                          { panic("implement me") }
func (s *upstartService) Install() error                                      { panic("implement me") }
func (s *upstartService) Uninstall() error                                    { panic("implement me") }
func (s *upstartService) GetLogger(errs chan<- error) (service.Logger, error) { panic("implement me") }
func (s *upstartService) SystemLogger(errs chan<- error) (service.Logger, error) {
	panic("implement me")
}
func (s *upstartService) String() string                  { return s.Name }
func (s *upstartService) Platform() string                { return "upstart" }
func (s *upstartService) Status() (service.Status, error) { panic("implement me") }
func (s *upstartService) Start() error                    { return runUpstartCommand("/sbin/start", s.Name) }
func (s *upstartService) Stop() error                     { return runUpstartCommand("/sbin/stop", s.Name) }
func (s *upstartService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}

func runUpstartCommand(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	return cmd.Run()
}

func IsUpstart() bool {
	_, err := exec.LookPath("initctl")
	return err == nil
}
