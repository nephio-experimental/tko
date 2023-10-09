package web

import (
	"net/http"

	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listSites(writer http.ResponseWriter, request *http.Request) {
	if sites, err := self.Backend.ListSites(nil, nil, nil); err == nil {
		sites_ := make([]ard.StringMap, len(sites))
		for index, site := range sites {
			sites_[index] = ard.StringMap{
				"id":          site.SiteID,
				"template":    site.TemplateID,
				"metadata":    site.Metadata,
				"deployments": site.DeploymentIDs,
			}
		}
		sortById(sites_)
		(&transcribe.Transcriber{Writer: writer}).WriteJSON(sites_)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) getSite(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if site, err := self.Backend.GetSite(id); err == nil {
		writeResources(writer, site.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
