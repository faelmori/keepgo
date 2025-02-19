package options

type Status byte

const (
	StatusUnknown Status = iota
	StatusRunning
	StatusStopped
)

const (
	OptionKeepAliveDefault     = true
	OptionRunAtLoadDefault     = false
	OptionUserServiceDefault   = false
	OptionSessionCreateDefault = false
	OptionLogOutputDefault     = false

	OptionRunAtLoad           = "RunAtLoad"
	OptionKeepAlive           = "KeepAlive"
	OptionUserService         = "UserService"
	OptionSessionCreate       = "SessionCreate"
	OptionLogOutput           = "LogOutput"
	OptionPrefix              = "Prefix"
	OptionPrefixDefault       = "application"
	OptionRunWait             = "RunWait"
	OptionReloadSignal        = "ReloadSignal"
	OptionPIDFile             = "PIDFile"
	OptionLimitNOFILE         = "LimitNOFILE"
	OptionRestart             = "Restart"
	OptionSuccessExitStatus   = "SuccessExitStatus"
	OptionSystemdScript       = "SystemdScript"
	OptionSysvScript          = "SysvScript"
	OptionRCSScript           = "RCSScript"
	OptionUpstartScript       = "UpstartScript"
	OptionLaunchdConfig       = "LaunchdConfig"
	OptionOpenRCScript        = "OpenRCScript"
	OptionLogDirectory        = "LogDirectory"
	OptionLogDirectoryDefault = "LogDirectoryDefault"

	OptionLimitNOFILEDefault = -1
)
