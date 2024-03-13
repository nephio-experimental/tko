package memory

import (
	contextpkg "context"
	"time"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *MemoryBackend) SetSite(context contextpkg.Context, site *backend.Site) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	var originalDeploymentIds []string
	if originalSite, ok := self.sites[site.SiteID]; ok {
		originalDeploymentIds = originalSite.DeploymentIDs
	}

	// Validate and merge template
	if site.TemplateID != "" {
		if template, ok := self.templates[site.TemplateID]; ok {
			site.MergeTemplate(template)
		} else {
			return backend.NewBadArgumentErrorf("unknown template: %s", site.TemplateID)
		}
	}

	// Restore associated deployments
	site.DeploymentIDs = originalDeploymentIds

	site.Updated = time.Now().UTC()
	self.sites[site.SiteID] = site

	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) GetSite(context contextpkg.Context, siteId string) (*backend.Site, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if site, ok := self.sites[siteId]; ok {
		return site.Clone(true), nil
	} else {
		return nil, backend.NewNotFoundErrorf("site: %s", siteId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeleteSite(context contextpkg.Context, siteId string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if _, ok := self.sites[siteId]; ok {
		delete(self.sites, siteId)

		// Remove deployment associations
		for _, deployment := range self.deployments {
			if deployment.SiteID == siteId {
				deployment.SiteID = ""
			}
		}

		return nil
	} else {
		return backend.NewNotFoundErrorf("site: %s", siteId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) ListSites(context contextpkg.Context, selectSites backend.SelectSites, window backend.Window) (util.Results[backend.SiteInfo], error) {
	self.lock.Lock()

	var siteInfos []backend.SiteInfo
	for _, site := range self.sites {
		if !backend.IDMatchesPatterns(site.TemplateID, selectSites.TemplateIDPatterns) {
			continue
		}

		if !backend.IDMatchesPatterns(site.SiteID, selectSites.SiteIDPatterns) {
			continue
		}

		if !backend.MetadataMatchesPatterns(site.Metadata, selectSites.MetadataPatterns) {
			continue
		}

		siteInfos = append(siteInfos, site.SiteInfo)
	}

	self.lock.Unlock()

	backend.SortSiteInfos(siteInfos)

	length := uint(len(siteInfos))
	if window.Offset > length {
		siteInfos = nil
	} else if end := window.Offset + window.MaxCount; end > length {
		siteInfos = siteInfos[window.Offset:]
	} else {
		siteInfos = siteInfos[window.Offset:end]
	}

	return util.NewResultsSlice(siteInfos), nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) PurgeSites(context contextpkg.Context, selectSites backend.SelectSites) error {
	return backend.NewNotImplementedError("PurgeSites")
}
