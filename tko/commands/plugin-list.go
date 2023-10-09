package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	pluginCommand.AddCommand(pluginListCommand)
}

var pluginListCommand = &cobra.Command{
	Use:   "list",
	Short: "List plugins",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListPlugins()
	},
}

func ListPlugins() {
	pluginInfos, err := NewClient().ListPlugins()
	FailOnGRPCError(err)
	Print(pluginInfos)
}
