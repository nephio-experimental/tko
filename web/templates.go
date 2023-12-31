package web

import (
	"net/http"

	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listTemplates(writer http.ResponseWriter, request *http.Request) {
	if templates, err := self.Backend.ListTemplates(nil, nil); err == nil {
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
		(&transcribe.Transcriber{Writer: writer}).WriteJSON(templates_)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) getTemplate(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if template, err := self.Backend.GetTemplate(id); err == nil {
		writeResources(writer, template.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
