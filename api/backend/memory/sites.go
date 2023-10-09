package memory

import (
	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
)

// ([backend.Backend] interface)
func (self *MemoryBackend) SetSite(site *backend.Site) error {
	site = site.Clone()
	if site.Metadata == nil {
		site.Metadata = make(map[string]string)
	}

	site.Update()

	self.lock.Lock()
	defer self.lock.Unlock()

	// TODO: merge template resources

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
func (self *MemoryBackend) GetSite(siteId string) (*backend.Site, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if site, ok := self.sites[siteId]; ok {
		return site.Clone(), nil
	} else {
		return nil, backend.NewNotFoundErrorf("site: %s", siteId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeleteSite(siteId string) error {
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
func (self *MemoryBackend) ListSites(siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]backend.SiteInfo, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	var siteInfos []backend.SiteInfo
	for _, site := range self.sites {
		if len(templateIdPatterns) > 0 {
			if !backend.IdMatchesPatterns(site.TemplateID, templateIdPatterns) {
				continue
			}
		}

		if backend.IdMatchesPatterns(site.SiteID, siteIdPatterns) && backend.MetadataMatchesPatterns(site.Metadata, metadataPatterns) {
			siteInfos = append(siteInfos, site.SiteInfo)
		}
	}

	return siteInfos, nil
}

// Utils

// Call when lock acquired
func (self *MemoryBackend) mergeSiteResources(site *backend.Site) {
	if template, ok := self.templates[site.TemplateID]; ok {
		resources := util.CopyResources(template.Resources)

		// Merge our resources over template resources
		resources = util.MergeResources(resources, site.Resources)

		site.Resources = resources
	}
}
