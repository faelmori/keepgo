package linux

import (
	"time"
)

type solarisService struct {
	Name string
}

func (s *solarisService) Start() error {
	return run("/usr/sbin/svcadm", "enable", s.Name)
}

func (s *solarisService) Stop() error {
	return run("/usr/sbin/svcadm", "disable", s.Name)
}

func (s *solarisService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}
