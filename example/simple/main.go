// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// simple does nothing except block while running the service.
package main

import (
	"log"

	keepgo "github.com/faelmori/keepgo/internal"
)

var logger keepgo.Logger

type program struct{}

func (p *program) Start(s keepgo.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	// Do work here
}
func (p *program) Stop(s keepgo.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {
	svcConfig := &keepgo.Config{
		Name:        "GoServiceExampleSimple",
		DisplayName: "Go Service Example",
		Description: "This is an example Go keepgo.",
	}

	prg := &program{}
	s, err := keepgo.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.GetLogger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
