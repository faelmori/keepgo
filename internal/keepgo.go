// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package keepgo

import (
	"errors"
	"fmt"
)

const (
	OptionKeepAlive            = "KeepAlive"
	OptionKeepAliveDefault     = true
	OptionRunAtLoad            = "RunAtLoad"
	OptionRunAtLoadDefault     = false
	OptionUserService          = "UserService"
	OptionUserServiceDefault   = false
	OptionSessionCreate        = "SessionCreate"
	OptionSessionCreateDefault = false
	OptionLogOutput            = "LogOutput"
	OptionLogOutputDefault     = false
	OptionPrefix               = "Prefix"
	OptionPrefixDefault        = "application"

	OptionRunWait            = "RunWait"
	OptionReloadSignal       = "ReloadSignal"
	OptionPIDFile            = "PIDFile"
	OptionLimitNOFILE        = "LimitNOFILE"
	OptionLimitNOFILEDefault = -1 // -1 = don't set in configuration
	OptionRestart            = "Restart"

	OptionSuccessExitStatus = "SuccessExitStatus"

	OptionSystemdScript = "SystemdScript"
	OptionSysvScript    = "SysvScript"
	OptionRCSScript     = "RCSScript"
	OptionUpstartScript = "UpstartScript"
	OptionLaunchdConfig = "LaunchdConfig"
	OptionOpenRCScript  = "OpenRCScript"

	OptionLogDirectory = "LogDirectory"
)

type Status byte

const (
	StatusUnknown Status = iota // Status is unable to be determined due to an error or it was not installed.
	StatusRunning
	StatusStopped
)

type Config struct {
	Name        string   // Required name of the service. No spaces suggested.
	DisplayName string   // Display name, spaces allowed.
	Description string   // Long description of service.
	UserName    string   // Run as username.
	Arguments   []string // Run with arguments.

	// Optional field to specify the executable for service.
	// If empty the current executable is used.
	Executable string

	// Array of service dependencies.
	// Not yet fully implemented on Linux or OS X:
	//  1. Support linux-SystemVard dependencies, just put each full line as the
	//     element of the string array, such as
	//     "After=network.target syslog.target"
	//     "Requires=syslog.target"
	//     Note, such lines will be directly appended into the [Unit] of
	//     the generated service config file, will not check their correctness.
	Dependencies []string

	// The following fields are not supported on Windows.
	WorkingDirectory string // Initial working directory.
	ChRoot           string

	// System specific options.
	Option KeyValue

	EnvVars map[string]string
}

var (
	SystemVar         System
	SystemVarRegistry []System
)

var (
	// ErrNameFieldRequired is returned when Config.Name is empty.
	ErrNameFieldRequired = errors.New("Config.Name field is required.")
	// ErrNoServiceSystemDetected is returned when no SystemVar was detected.
	ErrNoServiceSystemDetected = errors.New("No service SystemVar detected.")
	// ErrNotInstalled is returned when the service is not installed.
	ErrNotInstalled = errors.New("the service is not installed")
)

// New creates a new service based on a service interface and configuration.
func New(i Controller, c *Config) (Service, error) {
	if len(c.Name) == 0 {
		return nil, ErrNameFieldRequired
	}
	if SystemVar == nil {
		return nil, ErrNoServiceSystemDetected
	}
	return SystemVar.New(i, c)
}

type KeyValue map[string]interface{}

func (kv KeyValue) Bool(name string, defaultValue bool) bool {
	if v, found := kv[name]; found {
		if castValue, is := v.(bool); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) int(name string, defaultValue int) int {
	if v, found := kv[name]; found {
		if castValue, is := v.(int); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) String(name string, defaultValue string) string {
	if v, found := kv[name]; found {
		if castValue, is := v.(string); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) float64(name string, defaultValue float64) float64 {
	if v, found := kv[name]; found {
		if castValue, is := v.(float64); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) funcSingle(name string, defaultValue func()) func() {
	if v, found := kv[name]; found {
		if castValue, is := v.(func()); is {
			return castValue
		}
	}
	return defaultValue
}
func Platform() string {
	if SystemVar == nil {
		return ""
	}
	return SystemVar.String()
}
func Interactive() bool {
	if SystemVar == nil {
		return true
	}
	return SystemVar.Interactive()
}
func NewSystem() System {
	for _, choice := range SystemVarRegistry {
		if choice.Detect() == false {
			continue
		}
		return choice
	}
	return nil
}
func ChooseSystem(a ...System) {
	SystemVarRegistry = a
	SystemVar = NewSystem()
}
func ChosenSystem() System {
	return SystemVar
}
func AvailableSystems() []System {
	return SystemVarRegistry
}

type System interface {
	// String returns a description of the SystemVar.
	String() string

	// Detect returns true if the SystemVar is available to use.
	Detect() bool

	// Interactive returns false if running under the SystemVar service manager
	// and true otherwise.
	Interactive() bool

	// New creates a new service for this SystemVar.
	New(i Controller, c *Config) (Service, error)
}
type Controller interface {
	Start(s Service) error
	Stop(s Service) error
}
type Shutdowner interface {
	Controller
	Shutdown(s Service) error
}

// TODO: Add Configure to Service interface.

type Service interface {
	// Run should be called shortly after the program entry point.
	// After Controller.Stop has finished running, Run will stop blocking.
	// After Run stops blocking, the program must exit shortly after.
	Run() error

	// Start signals to the OS service manager the given service should start.
	Start() error

	// Stop signals to the OS service manager the given service should stop.
	Stop() error

	// Restart signals to the OS service manager the given service should stop then start.
	Restart() error

	// Install setups up the given service in the OS service manager. This may require
	// greater rights. Will return an error if it is already installed.
	Install() error

	// Uninstall removes the given service from the OS service manager. This may require
	// greater rights. Will return an error if the service is not present.
	Uninstall() error

	// Opens and returns a SystemVar logger. If the user program is running
	// interactively rather then as a service, the returned logger will write to
	// os.Stderr. If errs is non-nil errors will be sent on errs as well as
	// returned from Logger's functions.
	Logger(errs chan<- error) (Logger, error)

	// SystemLogger opens and returns a SystemVar logger. If errs is non-nil errors
	// will be sent on errs as well as returned from Logger's functions.
	SystemLogger(errs chan<- error) (Logger, error)

	// String displays the name of the service. The display name if present,
	// otherwise the name.
	String() string

	// Platform displays the name of the SystemVar that manages the service.
	// In most cases this will be the same as service.Platform().
	Platform() string

	// Status returns the current service status.
	Status() (Status, error)
}

var ControlAction = [5]string{"start", "stop", "restart", "install", "uninstall"}

func Control(s Service, action string) error {
	var err error
	switch action {
	case ControlAction[0]:
		err = s.Start()
	case ControlAction[1]:
		err = s.Stop()
	case ControlAction[2]:
		err = s.Restart()
	case ControlAction[3]:
		err = s.Install()
	case ControlAction[4]:
		err = s.Uninstall()
	default:
		err = fmt.Errorf("Unknown action %s", action)
	}
	if err != nil {
		return fmt.Errorf("Failed to %s %v: %v", action, s, err)
	}
	return nil
}

type Logger interface {
	Error(v ...interface{}) error
	Warning(v ...interface{}) error
	Info(v ...interface{}) error

	Errorf(format string, a ...interface{}) error
	Warningf(format string, a ...interface{}) error
	Infof(format string, a ...interface{}) error
}
