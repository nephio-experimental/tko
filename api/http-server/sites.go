package server

import (
	contextpkg "context"
	"io"
	"net/http"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

func (self *Server) listSites(writer http.ResponseWriter, request *http.Request) {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.BackendTimeout)
	defer cancel()

	if siteInfoStream, err := self.Backend.ListSites(context, backend.ListSites{}); err == nil {
		var sites []ard.StringMap
		for {
			if siteInfo, err := siteInfoStream.Next(); err == nil {
				sites = append(sites, ard.StringMap{
					"id":          siteInfo.SiteID,
					"template":    siteInfo.TemplateID,
					"metadata":    siteInfo.Metadata,
					"deployments": siteInfo.DeploymentIDs,
				})
			} else if err == io.EOF {
				break
			} else {
				writer.WriteHeader(500)
				return
			}
		}
		sortById(sites)
		transcribe.NewTranscriber().SetWriter(writer).WriteJSON(sites)
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
