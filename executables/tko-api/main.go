package main

import (
	"github.com/nephio-experimental/tko/executables/tko-api/commands"
	"github.com/tliron/kutil/util"
)

func main() {
	util.ExitOnSignals()
	commands.Execute()
	util.Exit(0)
}
