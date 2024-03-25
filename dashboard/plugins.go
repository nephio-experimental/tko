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
		SetTableHeader(table, "Type", "Name", "Executor", "Triggers")

		var pluginIds []client.PluginID
		util.IterateResults(pluginResults, func(plugin client.Plugin) error {
			pluginIds = append(pluginIds, plugin.PluginID)
			row := FindPluginRow(table, plugin.PluginID)
			self.SetPluginRow(table, row, &plugin)
			return nil
		})

		CleanTableRows(table, func(row int) bool {
			return ContainsPlugin(pluginIds, GetPluginRow(table, row))
		})
	}
}

func (self *Application) SetPluginRow(table *tview.Table, row int, plugin *client.Plugin) {
	triggers := make([]string, len(plugin.Triggers))
	for index, trigger := range plugin.Triggers {
		triggers[index] = trigger.ShortString()
	}

	pluginDetails := &PluginDetails{plugin.PluginID, self.client}
	table.SetCell(row, 0, tview.NewTableCell(plugin.PluginID.Type).SetReference(pluginDetails))
	table.SetCell(row, 1, tview.NewTableCell(plugin.PluginID.Name).SetReference(pluginDetails))
	table.SetCell(row, 2, tview.NewTableCell(plugin.Executor).SetReference(pluginDetails))
	table.SetCell(row, 3, tview.NewTableCell(strings.Join(triggers, "; ")).SetReference(pluginDetails))
}

func ContainsPlugin(pluginIds []client.PluginID, pluginId client.PluginID) bool {
	for _, pluginId_ := range pluginIds {
		if (pluginId.Type == pluginId_.Type) && (pluginId.Name == pluginId_.Name) {
			return true
		}
	}
	return false
}

func GetPluginRow(table *tview.Table, row int) client.PluginID {
	return table.GetCell(row, 0).GetReference().(*PluginDetails).pluginId
}

func FindPluginRow(table *tview.Table, pluginId client.PluginID) int {
	rowCount := table.GetRowCount()
	for row := 1; row < rowCount; row++ {
		pluginId_ := GetPluginRow(table, row)
		if (pluginId.Type == pluginId_.Type) && (pluginId.Name == pluginId_.Name) {
			return row
		}
	}
	return rowCount
}

//
// PluginDetails
//

type PluginDetails struct {
	pluginId client.PluginID
	client   *client.Client
}

// ([Details] interface)
func (self *PluginDetails) GetTitle() string {
	return "Plugin: " + self.pluginId.Type + " " + self.pluginId.Name
}

// ([Details] interface)
func (self *PluginDetails) GetText() string {
	if plugin, ok, err := self.client.GetPlugin(self.pluginId); err == nil {
		if ok {
			return ToYAML(plugin)
		} else {
			return ""
		}
	} else {
		return err.Error()
	}
}
