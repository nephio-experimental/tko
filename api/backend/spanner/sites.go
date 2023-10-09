package spanner

import (
	"github.com/nephio-experimental/tko/api/backend"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetSite(site *backend.Site) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetSite(siteId string) (*backend.Site, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeleteSite(siteId string) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListSites(siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]backend.SiteInfo, error) {
	return nil, nil
}
