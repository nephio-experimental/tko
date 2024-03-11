package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	pluginCommand.AddCommand(pluginGetCommand)
}

var pluginGetCommand = &cobra.Command{
	Use:   "get [TYPE] [NAME]",
	Short: "Get plugin",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		GetPlugin(args[0], args[1])
	},
}

func GetPlugin(type_ string, name string) {
	plugin, ok, err := NewClient().GetPlugin(client.NewPluginID(type_, name))
	FailOnGRPCError(err)
	if ok {
		Print(plugin)
	} else {
		util.Fail("not found")
	}
}
