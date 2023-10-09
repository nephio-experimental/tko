package main

import (
	"github.com/nephio-experimental/tko/tko/commands"
	"github.com/tliron/kutil/util"

	_ "github.com/tliron/commonlog/simple"
)

func main() {
	util.ExitOnSIGTERM()
	commands.Execute()
	util.Exit(0)
}
