package main

import (
	"github.com/nephio-experimental/tko/tko-instantiation-controller/commands"
	"github.com/tliron/kutil/util"
)

func main() {
	util.ExitOnSignals()
	commands.Execute()
	util.Exit(0)
}
