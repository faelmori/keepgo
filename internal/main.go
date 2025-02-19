package internal

import (
	"fmt"
	"github.com/faelmori/kbx/mods/logz"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type IKeepGo interface {
	Execute() error
	Module() string
	Alias() string
	ShortDescription() string
	LongDescription() string
	Usage() string
	Examples() []string
	Active() bool
	Command() *cobra.Command
}

type keepGo struct {
	parentCmdName string
	printBanner   bool
	certPath      string
	keyPath       string
	configPath    string
}

var StartCmd = &cobra.Command{
	Use:         "start",
	Annotations: getDescriptions([]string{"Start a service", "Start a service"}, os.Getenv("KEEPGO_PRINT_BANNER") == "true"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Service started")
	},
}
var StopCmd = &cobra.Command{
	Use:         "stop",
	Annotations: getDescriptions([]string{"Stop a service", "Stop a service"}, os.Getenv("KEEPGO_PRINT_BANNER") == "true"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Service stopped")
	},
}
var RestartCmd = &cobra.Command{
	Use:         "restart",
	Annotations: getDescriptions([]string{"Restart a service", "Restart a service"}, os.Getenv("KEEPGO_PRINT_BANNER") == "true"),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Service restarted")
	},
}

func (m *keepGo) Alias() string {
	return "keepgo"
}
func (m *keepGo) ShortDescription() string {
	return "KeepGo: The ultimate web server for effortless task execution, offering unmatched security, flexibility, and infinite possibilities."
}
func (m *keepGo) LongDescription() string {
	return `KeepGo: The ultimate daemonizer for effortless service management, offering unmatched flexibility, portability, and intuitiveness.
Inspired by [kardianos/service], KeepGo allows you to run virtually any process as a persistent service on the most popular operating systems in the world, including Windows, macOS, and Linux.
With KeepGo, you can configure and manage services without needing to stop and reinstall them, providing a seamless and efficient user experience.

Explore the infinite possibilities with KeepGo and elevate your service management to new heights.`
}
func (m *keepGo) Usage() string {
	return "keepgo [command] [args]"
}
func (m *keepGo) Examples() []string {
	return []string{"KeepGo [command] [args]", "KeepGo server start -p '8000' -b '0.0.0.0'"}
}
func (m *keepGo) Active() bool {
	return true
}
func (m *keepGo) Module() string {
	return "KeepGo"
}
func (m *keepGo) Execute() error {
	if keepGoErr := m.Command().Execute(); keepGoErr != nil {
		return logz.ErrorLog(keepGoErr.Error(), "KeepGo")
	} else {
		return nil
	}
}
func (m *keepGo) Command() *cobra.Command {
	cmdKg := &cobra.Command{
		Use:         m.Module(),
		Aliases:     []string{m.Alias(), "service", "daemonize", "kg"},
		Example:     m.concatenateExamples(),
		Annotations: getDescriptions([]string{m.LongDescription(), m.ShortDescription()}, m.printBanner),
		Run:         func(cmd *cobra.Command, args []string) { _ = logz.ErrorLog("No command specified.", "KeepGo") },
	}
	cmdKg.AddCommand(StartCmd)
	cmdKg.AddCommand(RestartCmd)
	cmdKg.AddCommand(StopCmd)
	SetUsageDefinition(cmdKg)
	return cmdKg
}
func (m *keepGo) SetParentCmdName(rtCmd string) { m.parentCmdName = rtCmd }
func (m *keepGo) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}

func getDescriptions(descriptionArg []string, _ bool) map[string]string {
	var description, banner string
	if descriptionArg != nil {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = descriptionArg[0]
		} else {
			description = descriptionArg[1]
		}
	}
	banner = `    __ __                ______    
   / //_/__  ___  ____  / ____/___ 
  / ,< / _ \/ _ \/ __ \/ / __/ __ \
 / /| /  __/  __/ /_/ / /_/ / /_/ /
/_/ |_\___/\___/ .___/\____/\____/ 
              /_/                  
`
	return map[string]string{"banner": banner, "description": description}
}

func NewKeepGo() IKeepGo {
	kg := keepGo{
		parentCmdName: "",
		printBanner:   os.Getenv("KEEPGO_PRINT_BANNER") == "true",
	}
	return &kg
}
