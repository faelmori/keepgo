package logging

import (
	"log"
	"os"
	"strings"
)

var ConsoleLoggerObj = ConsoleLoggerImpl{}

type ConsoleLoggerImpl struct {
	info, warn, err *log.Logger
}

func init() {
	ConsoleLoggerObj.info = log.New(os.Stderr, "I: ", log.Ltime)
	ConsoleLoggerObj.warn = log.New(os.Stderr, "W: ", log.Ltime)
	ConsoleLoggerObj.err = log.New(os.Stderr, "E: ", log.Ltime)
}

func (c ConsoleLoggerImpl) Error(v ...interface{}) error {
	c.err.Print(v...)
	return nil
}
func (c ConsoleLoggerImpl) Warning(v ...interface{}) error {
	c.warn.Print(v...)
	return nil
}
func (c ConsoleLoggerImpl) Info(v ...interface{}) error {
	c.info.Print(v...)
	return nil
}
func (c ConsoleLoggerImpl) Errorf(format string, a ...interface{}) error {
	c.err.Printf(format, a...)
	return nil
}
func (c ConsoleLoggerImpl) Warningf(format string, a ...interface{}) error {
	c.warn.Printf(format, a...)
	return nil
}
func (c ConsoleLoggerImpl) Infof(format string, a ...interface{}) error {
	c.info.Printf(format, a...)
	return nil
}

func GetDescriptions(descriptionArg []string, _ bool) map[string]string {
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
