package server

import (
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

func (self *Server) listSites(writer http.ResponseWriter, request *http.Request) {
	if siteInfoResults, err := self.Backend.ListSites(request.Context(), backend.ListSites{}); err == nil {
		var sites []ard.StringMap
		if err := util.IterateResults(siteInfoResults, func(siteInfo backend.SiteInfo) error {
			sites = append(sites, ard.StringMap{
				"id":          siteInfo.SiteID,
				"template":    siteInfo.TemplateID,
				"metadata":    siteInfo.Metadata,
				"deployments": siteInfo.DeploymentIDs,
			})
			return nil
		}); err != nil {
			writer.WriteHeader(500)
			return
		}

		sortById(sites)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(sites)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) getSite(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if site, err := self.Backend.GetSite(request.Context(), id); err == nil {
		writeResources(writer, site.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
