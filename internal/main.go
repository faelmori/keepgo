package keepgo

import (
	"os"
	"path/filepath"
)

func (c *Config) ExecPath() (string, error) {
	if len(c.Executable) != 0 {
		return filepath.Abs(c.Executable)
	}
	return os.Executable()
}
