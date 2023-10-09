package commands

import (
	"github.com/nephio-experimental/tko/api/client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	pluginCommand.AddCommand(pluginDeleteCommand)
}

var pluginDeleteCommand = &cobra.Command{
	Use:   "delete [TYPE] [GROUP] [VERSION] [KIND]",
	Short: "Delete plugin",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		DeletePlugin(args[0], args[1], args[2], args[3])
	},
}

func DeletePlugin(type_ string, group string, version string, kind string) {
	pluginId := client.NewPluginID(type_, tkoutil.NewGVK(group, version, kind))
	ok, reason, err := NewClient().DeletePlugin(pluginId)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("deleted plugin: %s", pluginId)
	} else {
		util.Fail(reason)
	}
}
