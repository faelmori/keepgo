package options

import (
	"os"
	"path/filepath"
)

type config struct {
	Name             string   // Required name of the service. No spaces suggested.
	DisplayName      string   // Display name, spaces allowed.
	Description      string   // Long description of service.
	UserName         string   // Run as username.
	Arguments        []string // Run with arguments.
	Executable       string
	Dependencies     []string
	WorkingDirectory string // Initial working directory.
	ChRoot           string
	Option           KeyValue
	EnvVars          map[string]string
}

func (c *config) GetName() string           { return c.Name }
func (c *config) GetDisplayName() string    { return c.DisplayName }
func (c *config) GetDescription() string    { return c.Description }
func (c *config) GetUserName() string       { return c.UserName }
func (c *config) GetArguments() []string    { return c.Arguments }
func (c *config) GetDependencies() []string { return c.Dependencies }
func (c *config) GetWorkingDirectory() string {
	return c.WorkingDirectory
}
func (c *config) GetChRoot() string   { return c.ChRoot }
func (c *config) GetOption() KeyValue { return c.Option }
func (c *config) GetEnvVars() map[string]string {
	return c.EnvVars
}

func (c *config) SetName(name string)               { c.Name = name }
func (c *config) SetDisplayName(displayName string) { c.DisplayName = displayName }
func (c *config) SetDescription(description string) { c.Description = description }
func (c *config) SetUserName(userName string)       { c.UserName = userName }
func (c *config) SetArguments(arguments []string)   { c.Arguments = arguments }
func (c *config) SetDependencies(dependencies []string) {
	c.Dependencies = dependencies
}
func (c *config) SetWorkingDirectory(workingDirectory string) {
	c.WorkingDirectory = workingDirectory
}
func (c *config) SetChRoot(chRoot string)   { c.ChRoot = chRoot }
func (c *config) SetOption(option KeyValue) { c.Option = option }
func (c *config) SetEnvVars(envVars map[string]string) {
	c.EnvVars = envVars
}
func (c *config) ExecPath() (string, error) {
	if len(c.Executable) != 0 {
		return filepath.Abs(c.Executable)
	}
	return os.Executable()
}

func NewConfig() Config {
	cfg := config{
		Arguments: make([]string, 0),
		Option:    make(KeyValue),
		EnvVars:   make(map[string]string),
	}
	return &cfg
}
