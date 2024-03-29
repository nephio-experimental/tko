package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/backend"
	"github.com/nephio-experimental/tko/plugins"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var (
	type_        string
	namePatterns []string
	listExecutor string
	trigger      string
)

func init() {
	pluginCommand.AddCommand(pluginListCommand)

	pluginListCommand.Flags().UintVar(&offset, "offset", 0, "fetch results starting at this offset")
	pluginListCommand.Flags().UintVar(&maxCount, "max-count", backend.DefaultMaxCount, "maximum number of results to fetch")
	pluginListCommand.Flags().StringVarP(&type_, "type", "t", "", "filter by plugin type")
	pluginListCommand.Flags().StringArrayVarP(&namePatterns, "name", "n", nil, "filter by plugin name pattern")
	pluginListCommand.Flags().StringVarP(&listExecutor, "executor", "e", "", "filter by plugin executor")
	pluginListCommand.Flags().StringVarP(&trigger, "trigger", "g", "", "filter by plugin trigger (\"group,version,kind\")")
}

var pluginListCommand = &cobra.Command{
	Use:   "list",
	Short: "List plugins",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListPlugins(offset, maxCount, type_, namePatterns, listExecutor, ParseTrigger(trigger))
	},
}

func ListPlugins(offset uint, maxCount uint, type_ string, namePatterns []string, executor string, trigger *tkoutil.GVK) {
	if !plugins.IsValidPluginType(type_, true) {
		util.Failf("plugin type must be %s: %s", plugins.PluginTypesDescription, type_)
	}

	pluginInfos, err := NewClient().ListPlugins(client.SelectPlugins{
		Type:         &type_,
		NamePatterns: namePatterns,
		Executor:     &executor,
		Trigger:      trigger,
	}, offset, int(maxCount))
	FailOnGRPCError(err)
	pluginInfos_, err := util.GatherResults(pluginInfos)
	util.FailOnError(err)
	Print(pluginInfos_)
}
