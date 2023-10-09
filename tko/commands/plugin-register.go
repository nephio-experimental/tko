package commands

import (
	"github.com/nephio-experimental/tko/api/client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var executor string
var properties map[string]string

func init() {
	pluginCommand.AddCommand(pluginRegisterCommand)

	pluginRegisterCommand.Flags().StringVarP(&executor, "executor", "e", "command", "plugin executor")
	pluginRegisterCommand.Flags().StringToStringVarP(&properties, "property", "r", nil, "plugin property")
}

var pluginRegisterCommand = &cobra.Command{
	Use:   "register [TYPE] [GROUP] [VERSION] [KIND] [ARGUMENT...]",
	Short: "Register plugin",
	Args:  cobra.MinimumNArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		RegisterPlugin(args[0], args[1], args[2], args[3], executor, args[4:], properties)
	},
}

func RegisterPlugin(type_ string, group string, version string, kind string, executor string, arguments []string, properties map[string]string) {
	pluginId := client.NewPluginID(type_, tkoutil.NewGVK(group, version, kind))
	ok, reason, err := NewClient().RegisterPlugin(pluginId, executor, arguments, properties)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("registered plugin: %s", pluginId)
	} else {
		util.Fail(reason)
	}
}
