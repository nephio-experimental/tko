package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/plugins"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	pluginCommand.AddCommand(pluginPurgeCommand)

	pluginPurgeCommand.Flags().StringVarP(&type_, "type", "t", "", "filter by plugin type")
	pluginPurgeCommand.Flags().StringArrayVarP(&namePatterns, "name", "n", nil, "filter by plugin name pattern")
	pluginPurgeCommand.Flags().StringVarP(&listExecutor, "executor", "e", "", "filter by plugin executor")
	pluginPurgeCommand.Flags().StringVarP(&trigger, "trigger", "g", "", "filter by plugin trigger (\"group,version,kind\")")
}

var pluginPurgeCommand = &cobra.Command{
	Use:   "purge",
	Short: "Purge plugins",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		PurgePlugins(type_, namePatterns, listExecutor, ParseTrigger(trigger))
	},
}

func PurgePlugins(type_ string, namePatterns []string, executor string, trigger *tkoutil.GVK) {
	if !plugins.IsValidPluginType(type_, true) {
		util.Failf("plugin type must be %s: %s", plugins.PluginTypesDescription, type_)
	}

	ok, reason, err := NewClient().PurgePlugins(client.SelectPlugins{
		Type:         &type_,
		NamePatterns: namePatterns,
		Executor:     &executor,
		Trigger:      trigger,
	})
	FailOnGRPCError(err)
	if !ok {
		util.Fail(reason)
	}
}
