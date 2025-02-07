package runners

type Runner interface {
	RunWithOutput(command string, arguments ...string) (int, string, error)
}
