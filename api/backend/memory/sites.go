package memory

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
)

// ([backend.Backend] interface)
func (self *MemoryBackend) SetSite(context contextpkg.Context, site *backend.Site) error {
	site = site.Clone()
	if site.Metadata == nil {
		site.Metadata = make(map[string]string)
	}

	site.Update()

	self.lock.Lock()
	defer self.lock.Unlock()

	if site.TemplateID != "" {
		if _, ok := self.templates[site.TemplateID]; !ok {
			return backend.NewBadArgumentErrorf("unknown template: %s", site.TemplateID)
		}
		self.mergeSiteResources(site)
	}

	self.sites[site.SiteID] = site

	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) GetSite(context contextpkg.Context, siteId string) (*backend.Site, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if site, ok := self.sites[siteId]; ok {
		return site.Clone(), nil
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
func (self *MemoryBackend) ListSites(context contextpkg.Context, listSites backend.ListSites) ([]backend.SiteInfo, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	var siteInfos []backend.SiteInfo
	for _, site := range self.sites {
		if !backend.IDMatchesPatterns(site.TemplateID, listSites.TemplateIDPatterns) {
			continue
		}

		if !backend.IDMatchesPatterns(site.SiteID, listSites.SiteIDPatterns) {
			continue
		}

		if !backend.MetadataMatchesPatterns(site.Metadata, listSites.MetadataPatterns) {
			continue
		}

		siteInfos = append(siteInfos, site.SiteInfo)
	}

	return siteInfos, nil
}

// Utils

// Call when lock acquired
func (self *MemoryBackend) mergeSiteResources(site *backend.Site) {
	if template, ok := self.templates[site.TemplateID]; ok {
		site.MergeTemplate(template)
	}
}
