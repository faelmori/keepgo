package linux

import (
	"fmt"
	"github.com/faelmori/keepgo/service"
	"log/syslog"
	"strings"
	"text/template"
)

func NewSysLogger(name string, errs chan<- error) (service.Logger, error) {
	w, err := syslog.New(syslog.LOG_INFO, name)
	if err != nil {
		return nil, err
	}
	return sysLogger{w, errs}, nil
}

type sysLogger struct {
	*syslog.Writer
	errs chan<- error
}

func (s sysLogger) send(err error) error {
	if err != nil && s.errs != nil {
		s.errs <- err
	}
	return err
}
func (s sysLogger) Error(v ...interface{}) error {
	return s.send(s.Writer.Err(fmt.Sprint(v...)))
}
func (s sysLogger) Warning(v ...interface{}) error {
	return s.send(s.Writer.Warning(fmt.Sprint(v...)))
}
func (s sysLogger) Info(v ...interface{}) error {
	return s.send(s.Writer.Info(fmt.Sprint(v...)))
}
func (s sysLogger) Errorf(format string, a ...interface{}) error {
	return s.send(s.Writer.Err(fmt.Sprintf(format, a...)))
}
func (s sysLogger) Warningf(format string, a ...interface{}) error {
	return s.send(s.Writer.Warning(fmt.Sprintf(format, a...)))
}
func (s sysLogger) Infof(format string, a ...interface{}) error {
	return s.send(s.Writer.Info(fmt.Sprintf(format, a...)))
}

var TF = template.FuncMap{"cmdEscape": cmdEscape}

func cmdEscape(s string) string { return strings.ReplaceAll(s, " ", "\\ ") }
