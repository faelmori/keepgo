package linux

import (
	"bufio"
	"fmt"
	"github.com/faelmori/keepgo/service"
	"os"
	"os/exec"
	"strings"
	"time"
)

var CgroupFile = "/proc/1/cgroup"

type linuxSystemService struct {
	name        string
	detect      func() bool
	interactive func() bool
	new         func(i service.Controller, platform string, c *service.Config) (service.Service, error)
}

func (sc linuxSystemService) String() string {
	return sc.name
}
func (sc linuxSystemService) Detect() bool {
	return sc.detect()
}
func (sc linuxSystemService) Interactive() bool {
	return sc.interactive()
}
func (sc linuxSystemService) New(i service.Controller, c *service.Config) (service.Service, error) {
	return sc.new(i, sc.String(), c)
}

func init() {
	service.ChooseSystem(linuxSystemService{
		name:   "linux-systemd",
		detect: isSystemd,
		interactive: func() bool {
			is, _ := IsInteractive()
			return is
		},
		new: newSystemdService,
	},
		linuxSystemService{
			name:   "linux-upstart",
			detect: IsUpstart,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: newUpstartService,
		},
		linuxSystemService{
			name:   "linux-openrc",
			detect: service.Op.IsOpenRC,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: newOpenRCService,
		},
		linuxSystemService{
			name:   "linux-rcs",
			detect: isRCS,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: newRCSService,
		},
		linuxSystemService{
			name:   "unix-systemv",
			detect: func() bool { return true },
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: newSystemVService,
		},
	)
}

func BinaryName(pid int) (string, error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	dataBytes, err := os.ReadFile(statPath)
	if err != nil {
		return "", err
	}

	data := string(dataBytes)
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')
	return data[binStart : binStart+binEnd], nil
}

func IsInteractive() (bool, error) {
	inContainer, err := IsInContainer(CgroupFile)
	if err != nil {
		return false, err
	}

	if inContainer {
		return true, nil
	}

	ppid := os.Getppid()
	if ppid == 1 {
		return false, nil
	}

	binary, _ := BinaryName(ppid)
	return binary != "systemd", nil
}

func IsInContainer(cgroupPath string) (bool, error) {
	const maxlines = 5

	f, err := os.Open(cgroupPath)
	if err != nil {
		return false, err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)

	lines := 0
	for scan.Scan() && !(lines > maxlines) {
		if strings.Contains(scan.Text(), "docker") || strings.Contains(scan.Text(), "lxc") {
			return true, nil
		}
		lines++
	}
	if err := scan.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func run(command string, arguments ...string) error {
	cmd := exec.Command(command, arguments...)
	return cmd.Run()
}

type linuxService struct {
	Name string
}

func (s *linuxService) Start() error {
	return run("/bin/systemctl", "start", s.Name)
}

func (s *linuxService) Stop() error {
	return run("/bin/systemctl", "stop", s.Name)
}

func (s *linuxService) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return s.Start()
}
