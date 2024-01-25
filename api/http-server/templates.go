package server

import (
	contextpkg "context"
	"net/http"

	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listTemplates(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if templates, err := self.Backend.ListTemplates(context, nil, nil); err == nil {
		templates_ := make([]ard.StringMap, len(templates))
		for index, template := range templates {
			templates_[index] = ard.StringMap{
				"id":          template.TemplateID,
				"template":    template.TemplateID,
				"metadata":    template.Metadata,
				"deployments": template.DeploymentIDs,
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
