package dashboard

import (
	"slices"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdatePlugins(table *tview.Table) {
	if plugins, err := self.client.ListPlugins(client.ListPlugins{}); err == nil {
		if plugins_, err := util.GatherResults(plugins); err == nil {
			slices.SortFunc(plugins_, func(a client.Plugin, b client.Plugin) int {
				return strings.Compare(a.Type+"|"+a.Name, b.Type+"|"+b.Name)
			})

			table.Clear()

			SetTableHeader(table, "Type", "Name", "Executor", "Triggers")

			for row, plugin := range plugins_ {
				triggers := make([]string, len(plugin.Triggers))
				for index, trigger := range plugin.Triggers {
					triggers[index] = trigger.ShortString()
				}

				row++
				table.SetCellSimple(row, 0, plugin.PluginID.Type)
				table.SetCellSimple(row, 1, plugin.PluginID.Name)
				table.SetCellSimple(row, 2, plugin.Executor)
				table.SetCellSimple(row, 3, strings.Join(triggers, "; "))
			}
		}
	}
}
