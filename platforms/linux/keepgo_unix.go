package linux

import (
	"os/exec"
)

func runUnixCommand(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	return cmd.Run()
}
