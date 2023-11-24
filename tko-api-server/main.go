package main

import (
	"github.com/nephio-experimental/tko/tko-api-server/commands"
	"github.com/tliron/kutil/util"
)

func main() {
	util.ExitOnSignals()
	commands.Execute()
	util.Exit(0)
}
