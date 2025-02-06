// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// simple does nothing except block while running the service.
package main

import (
	"log"
	"os"
	"time"

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
	<-time.After(time.Second * 13)
	return nil
}

func main() {
	svcConfig := &keepgo.Config{
		Name:        "GoServiceExampleStopPause",
		DisplayName: "Go Service Example: Stop Pause",
		Description: "This is an example Go keepgo that pauses on stop.",
	}

	prg := &program{}
	s, err := keepgo.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		err = keepgo.Control(s, os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		return
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
