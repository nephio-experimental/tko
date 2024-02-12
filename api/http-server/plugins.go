package server

import (
	contextpkg "context"
	"io"
	"net/http"

	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listPlugins(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if pluginStream, err := self.Backend.ListPlugins(context); err == nil {
		var plugins []ard.StringMap
		for {
			if plugin, err := pluginStream.Next(); err == nil {
				plugins = append(plugins, ard.StringMap{
					"id":         plugin.Type + "|" + plugin.APIVersion() + "|" + plugin.Kind,
					"type":       plugin.Type,
					"gvk":        []string{plugin.APIVersion(), plugin.Kind},
					"executor":   plugin.Executor,
					"arguments":  plugin.Arguments,
					"properties": plugin.Properties,
				})
			} else if err == io.EOF {
				break
			} else {
				writer.WriteHeader(500)
				return
			}
		}
		sortById(plugins)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(plugins)
	} else {
		writer.WriteHeader(500)
	}
}
