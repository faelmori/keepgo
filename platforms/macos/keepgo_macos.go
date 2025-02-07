package macos

import (
	"fmt"
	"os/exec"
	"time"
)

type macosService struct {
	Name any
}

func (s *macosService) Start() error {
	return run("/bin/launchctl", "load", s.getPlistPath())
}

func (s *macosService) Stop() error {
	return run("/bin/launchctl", "unload", s.getPlistPath())
}

func (s *macosService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}

func (s *macosService) getPlistPath() string {
	return fmt.Sprintf("/Library/LaunchDaemons/%s.plist", s.Name)
}

func run(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	return cmd.Run()
}
