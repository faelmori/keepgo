package linux

import (
	"github.com/faelmori/keepgo/service"
	"os/exec"
	"time"
)

func NewOpenRCService(i service.Controller, platform string, c *service.Config) (service.Service, error) {
	return &openRCService{Name: c.Name}, nil
}

type openRCService struct{ Name string }

func (s *openRCService) Run() error                                          { panic("implement me") }
func (s *openRCService) Install() error                                      { panic("implement me") }
func (s *openRCService) Uninstall() error                                    { panic("implement me") }
func (s *openRCService) GetLogger(errs chan<- error) (service.Logger, error) { panic("implement me") }
func (s *openRCService) SystemLogger(errs chan<- error) (service.Logger, error) {
	panic("implement me")
}
func (s *openRCService) String() string                  { panic("implement me") }
func (s *openRCService) Platform() string                { panic("implement me") }
func (s *openRCService) Status() (service.Status, error) { panic("implement me") }
func (s *openRCService) Start() error                    { return run("/etc/init.d/"+s.Name, "start") }
func (s *openRCService) Stop() error                     { return run("/etc/init.d/"+s.Name, "stop") }

func (s *openRCService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}

func IsOpenRC() bool {
	_, err := exec.LookPath("openrc")
	return err == nil
}
