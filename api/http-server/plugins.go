package server

import (
	"net/http"
	"slices"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) ListPlugins(writer http.ResponseWriter, request *http.Request) {
	if pluginResults, err := self.Backend.ListPlugins(request.Context(), backend.SelectPlugins{}, backend.Window{}); err == nil {
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
			writer.WriteHeader(500)
			return
		}

		sortById(plugins)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(plugins)
	} else {
		writer.WriteHeader(500)
	}
}
