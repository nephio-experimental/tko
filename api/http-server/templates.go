package server

import (
	contextpkg "context"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listTemplates(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if templateInfoStream, err := self.Backend.ListTemplates(context, backend.ListTemplates{}); err == nil {
		var templates_ []ard.StringMap
		for {
			if templateInfo, ok := templateInfoStream.Next(); ok {
				templates_ = append(templates_, ard.StringMap{
					"id":          templateInfo.TemplateID,
					"template":    templateInfo.TemplateID,
					"metadata":    templateInfo.Metadata,
					"deployments": templateInfo.DeploymentIDs,
				})
			} else {
				break
			}
		}
		sortById(templates_)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(templates_)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) getTemplate(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	id := request.URL.Query().Get("id")
	if template, err := self.Backend.GetTemplate(context, id); err == nil {
		writeResources(writer, template.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
