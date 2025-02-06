// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package cmd

import (
	"log"
	"os"
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
