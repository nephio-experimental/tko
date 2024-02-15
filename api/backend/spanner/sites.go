package spanner

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/kutil/util"
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
func (self *SpannerBackend) ListSites(context contextpkg.Context, listSites backend.ListSites) (util.Results[backend.SiteInfo], error) {
	return nil, nil
}
