package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	pluginCommand.AddCommand(pluginDeleteCommand)
}

var pluginDeleteCommand = &cobra.Command{
	Use:   "delete [TYPE] [NAME]",
	Short: "Delete plugin",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		DeletePlugin(args[0], args[1])
	},
}

func DeletePlugin(type_ string, name string) {
	pluginId := client.NewPluginID(type_, name)
	ok, reason, err := NewClient().DeletePlugin(pluginId)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("deleted plugin: %s", pluginId)
	} else {
		util.Fail(reason)
	}
}
