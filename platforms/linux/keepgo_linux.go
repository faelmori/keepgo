package linux

import (
	"bufio"
	"fmt"
	"github.com/faelmori/keepgo/internal/linux"
	"github.com/faelmori/keepgo/service"
	"os"
	"strings"
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
	var systems []linuxSystemService

	if linux.IsSystemd() {
		systems = append(systems, linuxSystemService{
			name:   "linux-systemd",
			detect: linux.IsSystemd,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: linux.NewSystemdService,
		})
	}
	if linux.IsUpstart() {
		systems = append(systems, linuxSystemService{
			name:   "linux-upstart",
			detect: linux.IsUpstart,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: linux.NewUpstartService,
		})
	}
	if linux.IsOpenRC() {
		systems = append(systems, linuxSystemService{
			name:   "linux-openrc",
			detect: linux.IsOpenRC,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: linux.NewOpenRCService,
		})
	}
	if linux.IsRCS() {
		systems = append(systems, linuxSystemService{
			name:   "linux-rcs",
			detect: linux.IsRCS,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: linux.NewRCSService,
		})
	}
	if linux.IsUnix() {
		systems = append(systems, linuxSystemService{
			name:   "unix",
			detect: linux.IsUnix,
			interactive: func() bool {
				is, _ := IsInteractive()
				return is
			},
			new: linux.NewUnixService,
		})
	}

	// Passa apenas os sistemas detectados
	service.ChooseSystem(systems...)
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
