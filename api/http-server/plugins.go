package server

import (
	"net/http"
	"slices"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

func (self *Server) ListPlugins(writer http.ResponseWriter, request *http.Request) {
	// TODO: paging
	if pluginResults, err := self.Backend.ListPlugins(request.Context(), backend.SelectPlugins{}, getWindow(request)); err == nil {
		var plugins []ard.StringMap
		if err := util.IterateResults(pluginResults, func(plugin backend.Plugin) error {
			triggers := make([]string, len(plugin.Triggers))
			for index, trigger := range plugin.Triggers {
				triggers[index] = trigger.ShortString()
			}
			slices.Sort(triggers)

			plugins = append(plugins, ard.StringMap{
				"id":         plugin.PluginID.String(),
				"type":       plugin.Type,
				"name":       plugin.Name,
				"executor":   plugin.Executor,
				"arguments":  plugin.Arguments,
				"properties": plugin.Properties,
				"triggers":   triggers,
			})

			return nil
		}); err != nil {
			self.error(writer, err)
			return
		}

		self.writeJson(writer, plugins)
	} else {
		self.error(writer, err)
	}
}
