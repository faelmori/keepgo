// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package service

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
	OptionSuccessExitStatus  = "SuccessExitStatus"

	OptionSystemdScript = "SystemdScript"
	OptionSysvScript    = "SysvScript"
	OptionRCSScript     = "RCSScript"
	OptionUpstartScript = "UpstartScript"
	OptionLaunchdConfig = "LaunchdConfig"
	OptionOpenRCScript  = "OpenRCScript"
	OptionLogDirectory  = "LogDirectory"

	StatusUnknown Status = iota // Status is unable to be determined due to an error or it was not installed.
	StatusRunning
	StatusStopped
)

var (
	SystemVar         System
	SystemVarRegistry []System

	// ErrNameFieldRequired is returned when Config.Name is empty.
	ErrNameFieldRequired = errors.New("Config.Name field is required.")
	// ErrNoServiceSystemDetected is returned when no SystemVar was detected.
	ErrNoServiceSystemDetected = errors.New("No service SystemVar detected.")
	// ErrNotInstalled is returned when the service is not installed.
	ErrNotInstalled = errors.New("the service is not installed")
	ControlAction   = [5]string{"start", "stop", "restart", "install", "uninstall"}
)

type Status byte
type Config struct {
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
type KeyValue map[string]interface{}
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
type Service interface {
	Run() error
	Start() error
	Stop() error
	Restart() error
	Install() error
	Uninstall() error
	GetLogger(errs chan<- error) (Logger, error)
	SystemLogger(errs chan<- error) (Logger, error)
	String() string
	Platform() string
	Status() (Status, error)
}
type Logger interface {
	Error(v ...interface{}) error
	Warning(v ...interface{}) error
	Info(v ...interface{}) error

	Errorf(format string, a ...interface{}) error
	Warningf(format string, a ...interface{}) error
	Infof(format string, a ...interface{}) error
}

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
func (kv KeyValue) Bool(name string, defaultValue bool) bool {
	if v, found := kv[name]; found {
		if castValue, is := v.(bool); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) Int(name string, defaultValue int) int {
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
func (kv KeyValue) Float64(name string, defaultValue float64) float64 {
	if v, found := kv[name]; found {
		if castValue, is := v.(float64); is {
			return castValue
		}
	}
	return defaultValue
}
func (kv KeyValue) FuncSingle(name string, defaultValue func()) func() {
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

// TODO: Add Configure to Service interface.

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
