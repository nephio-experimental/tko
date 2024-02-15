package server

import (
	contextpkg "context"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) listTemplates(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if templateInfoResults, err := self.Backend.ListTemplates(context, backend.ListTemplates{}); err == nil {
		var templates []ard.StringMap
		if err := util.IterateResults(templateInfoResults, func(templateInfo backend.TemplateInfo) error {
			templates = append(templates, ard.StringMap{
				"id":          templateInfo.TemplateID,
				"metadata":    templateInfo.Metadata,
				"deployments": templateInfo.DeploymentIDs,
			})
			return nil
		}); err != nil {
			writer.WriteHeader(500)
			return
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
