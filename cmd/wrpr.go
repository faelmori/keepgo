package cmd

import (
	i "github.com/faelmori/keepgo/internal"
)

var keepGo = i.NewKeepGo()

func ExecuteKeepGo() error {
	return keepGo.Execute()
}
