package spanner

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetSite(context contextpkg.Context, site *backend.Site) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetSite(context contextpkg.Context, siteId string) (*backend.Site, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeleteSite(context contextpkg.Context, siteId string) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListSites(context contextpkg.Context, siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]backend.SiteInfo, error) {
	return nil, nil
}
