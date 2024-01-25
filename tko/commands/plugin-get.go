package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	pluginCommand.AddCommand(pluginGetCommand)
}

var pluginGetCommand = &cobra.Command{
	Use:   "get [TYPE] [GROUP] [VERSION] [KIND]",
	Short: "Get plugin",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		GetPlugin(args[0], args[1], args[2], args[3])
	},
}

func GetPlugin(type_ string, group string, version string, kind string) {
	pluginId := client.NewPluginID(type_, tkoutil.NewGVK(group, version, kind))
	plugin, ok, err := NewClient().GetPlugin(pluginId)
	FailOnGRPCError(err)
	if ok {
		Print(plugin)
	} else {
		util.Fail("not found")
	}
}
