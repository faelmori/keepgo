package main

import (
	"github.com/faelmori/kbx/mods/logz"
	"github.com/faelmori/keepgo/cmd"
)

func main() {
	if err := cmd.ExecuteKeepGo(); err != nil {
		logz.Panic(err)
	}
}
