package linux

import (
	"github.com/faelmori/keepgo/service"
	"os/exec"
	"time"
)

func NewRCSService(i service.Controller, platform string, c *service.Config) (service.Service, error) {
	return &rcsService{Name: c.Name}, nil
}

type rcsService struct {
	Name string
}

func (s *rcsService) Run() error                                          { panic("implement me") }
func (s *rcsService) Install() error                                      { panic("implement me") }
func (s *rcsService) Uninstall() error                                    { panic("implement me") }
func (s *rcsService) GetLogger(errs chan<- error) (service.Logger, error) { panic("implement me") }
func (s *rcsService) SystemLogger(errs chan<- error) (service.Logger, error) {
	panic("implement me")
}
func (s *rcsService) String() string                  { return s.Name }
func (s *rcsService) Platform() string                { return "rcs" }
func (s *rcsService) Status() (service.Status, error) { panic("implement me") }
func (s *rcsService) Start() error                    { return runRcsCommand("/etc/rc.d/"+s.Name, "start") }
func (s *rcsService) Stop() error                     { return runRcsCommand("/etc/rc.d/"+s.Name, "stop") }
func (s *rcsService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}

func runRcsCommand(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	return cmd.Run()
}

func IsRCS() bool {
	_, err := exec.LookPath("rc-status")
	return err == nil
}
