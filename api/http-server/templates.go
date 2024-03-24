package server

import (
	"net/http"
	"slices"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

func (self *Server) ListTemplates(writer http.ResponseWriter, request *http.Request) {
	// TODO: paging
	if templateInfoResults, err := self.Backend.ListTemplates(request.Context(), backend.SelectTemplates{}, getWindow(request)); err == nil {
		var templates []ard.StringMap
		if err := util.IterateResults(templateInfoResults, func(templateInfo backend.TemplateInfo) error {
			slices.Sort(templateInfo.DeploymentIDs)
			templates = append(templates, ard.StringMap{
				"id":               templateInfo.TemplateID,
				"deployments":      templateInfo.DeploymentIDs,
				"metadata":         templateInfo.Metadata,
				"updatedTimestamp": self.timestamp(templateInfo.Updated),
			})
			return nil
		}); err != nil {
			self.error(writer, err)
			return
		}

		self.writeJson(writer, templates)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) GetTemplate(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if template, err := self.Backend.GetTemplate(request.Context(), id); err == nil {
		self.writePackage(writer, template.Package)
	} else {
		self.error(writer, err)
	}
}
