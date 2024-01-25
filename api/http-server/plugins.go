package server

import (
	contextpkg "context"
	"net/http"

	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listPlugins(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if plugins, err := self.Backend.ListPlugins(context); err == nil {
		plugins_ := make([]ard.StringMap, len(plugins))
		for index, plugin := range plugins {
			plugins_[index] = ard.StringMap{
				"id":         plugin.Type + "|" + plugin.APIVersion() + "|" + plugin.Kind,
				"type":       plugin.Type,
				"gvk":        []string{plugin.APIVersion(), plugin.Kind},
				"executor":   plugin.Executor,
				"arguments":  plugin.Arguments,
				"properties": plugin.Properties,
			}
		}
		sortById(plugins_)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(plugins_)
	} else {
		writer.WriteHeader(500)
	}
}
