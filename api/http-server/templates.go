package server

import (
	contextpkg "context"
	"io"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listTemplates(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if templateInfoStream, err := self.Backend.ListTemplates(context, backend.ListTemplates{}); err == nil {
		var templates []ard.StringMap
		for {
			if templateInfo, err := templateInfoStream.Next(); err == nil {
				templates = append(templates, ard.StringMap{
					"id":          templateInfo.TemplateID,
					"template":    templateInfo.TemplateID,
					"metadata":    templateInfo.Metadata,
					"deployments": templateInfo.DeploymentIDs,
				})
			} else if err == io.EOF {
				break
			} else {
				writer.WriteHeader(500)
				return
			}
		}
		sortById(templates)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(templates)
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
