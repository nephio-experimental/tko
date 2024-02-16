package commands

import (
	"github.com/nephio-experimental/tko/dashboard"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	rootCommand.AddCommand(tuiCommand)
}

var tuiCommand = &cobra.Command{
	Use:   "dashboard",
	Short: "Start dashboard TUI",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := dashboard.Dashboard(NewClient())
		util.FailOnError(err)
	},
}
