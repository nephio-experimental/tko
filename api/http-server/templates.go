package server

import (
	"net/http"
	"slices"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) ListTemplates(writer http.ResponseWriter, request *http.Request) {
	if templateInfoResults, err := self.Backend.ListTemplates(request.Context(), backend.ListTemplates{}); err == nil {
		var templates []ard.StringMap
		if err := util.IterateResults(templateInfoResults, func(templateInfo backend.TemplateInfo) error {
			slices.Sort(templateInfo.DeploymentIDs)
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

func (self *Server) GetTemplate(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if template, err := self.Backend.GetTemplate(request.Context(), id); err == nil {
		writeResources(writer, template.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
