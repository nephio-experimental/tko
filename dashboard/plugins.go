package dashboard

import (
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdatePlugins(table *tview.Table) {
	// TODO: paging
	if pluginResults, err := self.client.ListPlugins(client.SelectPlugins{}, 0, -1); err == nil {
		table.Clear()

		SetTableHeader(table, "Type", "Name", "Executor", "Triggers")

		row := 1
		util.IterateResults(pluginResults, func(plugin client.Plugin) error {
			triggers := make([]string, len(plugin.Triggers))
			for index, trigger := range plugin.Triggers {
				triggers[index] = trigger.ShortString()
			}

			table.SetCellSimple(row, 0, plugin.PluginID.Type)
			table.SetCellSimple(row, 1, plugin.PluginID.Name)
			table.SetCellSimple(row, 2, plugin.Executor)
			table.SetCellSimple(row, 3, strings.Join(triggers, "; "))

			row++
			return nil
		})
	}
}
