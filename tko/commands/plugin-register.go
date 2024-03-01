package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var (
	properties map[string]string
	triggers   []string
)

func init() {
	pluginCommand.AddCommand(pluginRegisterCommand)

	pluginRegisterCommand.Flags().StringVarP(&executor, "executor", "e", "command", "plugin executor")
	pluginRegisterCommand.Flags().StringToStringVarP(&properties, "property", "r", nil, "executor property")
	pluginRegisterCommand.Flags().StringArrayVarP(&triggers, "trigger", "g", nil, "plugin trigger (\"group,version,kind\")")
}

var pluginRegisterCommand = &cobra.Command{
	Use:   "register [TYPE] [NAME] [ARGUMENT...]",
	Short: "Register plugin",
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		var triggers_ []tkoutil.GVK
		for _, trigger := range triggers {
			if gvk := ParseTrigger(trigger); gvk != nil {
				triggers_ = append(triggers_, *gvk)
			}
		}

		RegisterPlugin(args[0], args[1], executor, args[2:], properties, triggers_)
	},
}

func RegisterPlugin(type_ string, name string, executor string, arguments []string, properties map[string]string, triggers []tkoutil.GVK) {
	if !tkoutil.IsValidPluginType(type_, false) {
		util.Failf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, type_)
	}

	pluginId := client.NewPluginID(type_, name)
	ok, reason, err := NewClient().RegisterPlugin(pluginId, executor, arguments, properties, triggers)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("registered plugin: %s", pluginId)
	} else {
		util.Fail(reason)
	}
}
