package commands

import (
	"github.com/nephio-experimental/tko/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(tuiCommand)
}

var tuiCommand = &cobra.Command{
	Use:   "tui",
	Short: "Start TUI (Terminal User Interface)",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		tui.Start()
	},
}
