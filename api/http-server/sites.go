package server

import (
	"net/http"
	"slices"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

func (self *Server) ListSites(writer http.ResponseWriter, request *http.Request) {
	// TODO: paging
	if siteInfoResults, err := self.Backend.ListSites(request.Context(), backend.SelectSites{}, getWindow(request)); err == nil {
		var sites []ard.StringMap
		if err := util.IterateResults(siteInfoResults, func(siteInfo backend.SiteInfo) error {
			slices.Sort(siteInfo.DeploymentIDs)
			sites = append(sites, ard.StringMap{
				"id":               siteInfo.SiteID,
				"template":         siteInfo.TemplateID,
				"deployments":      siteInfo.DeploymentIDs,
				"metadata":         siteInfo.Metadata,
				"updatedTimestamp": self.timestamp(siteInfo.Updated),
			})
			return nil
		}); err != nil {
			self.error(writer, err)
			return
		}

		self.writeJson(writer, sites)
	} else {
		self.error(writer, err)
	}
}

func (self *Server) GetSite(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	if site, err := self.Backend.GetSite(request.Context(), id); err == nil {
		self.writePackage(writer, site.Package)
	} else {
		self.error(writer, err)
	}
}
