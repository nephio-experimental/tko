package server

import (
	contextpkg "context"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) listPlugins(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if pluginResults, err := self.Backend.ListPlugins(context, backend.ListPlugins{}); err == nil {
		var plugins []ard.StringMap
		if err := util.IterateResults(pluginResults, func(plugin backend.Plugin) error {
			triggers := make([]string, len(plugin.Triggers))
			for index, trigger := range plugin.Triggers {
				triggers[index] = trigger.ShortString()
			}

			plugins = append(plugins, ard.StringMap{
				"id":         plugin.Type + "|" + plugin.Name,
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
