package main

import (
	"github.com/nephio-experimental/tko/tko-preparation-controller/commands"
	"github.com/tliron/kutil/util"
)

func main() {
	util.ExitOnSIGTERM()
	commands.Execute()
	util.Exit(0)
}
