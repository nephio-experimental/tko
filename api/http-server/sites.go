package server

import (
	contextpkg "context"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listSites(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if sites, err := self.Backend.ListSites(context, backend.ListSites{}); err == nil {
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
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(sites_)
	} else {
		writer.WriteHeader(500)
	}
}

func (self *Server) getSite(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	id := request.URL.Query().Get("id")
	if site, err := self.Backend.GetSite(context, id); err == nil {
		writeResources(writer, site.Resources)
	} else {
		writer.WriteHeader(500)
	}
}
