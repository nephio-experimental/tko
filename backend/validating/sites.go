package validating

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *ValidatingBackend) SetSite(context contextpkg.Context, site *backend.Site) error {
	if site.SiteID == "" {
		return backend.NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(site.SiteID) {
		return backend.NewBadArgumentError("invalid siteId")
	}

	if err := self.Validation.ValidatePackage(site.Package, true); err != nil {
		return backend.WrapBadArgumentError(err)
	}

	return self.Backend.SetSite(context, site)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) GetSite(context contextpkg.Context, siteId string) (*backend.Site, error) {
	if siteId == "" {
		return nil, backend.NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(siteId) {
		return nil, backend.NewBadArgumentError("invalid siteId")
	}

	return self.Backend.GetSite(context, siteId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) DeleteSite(context contextpkg.Context, siteId string) error {
	if siteId == "" {
		return backend.NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(siteId) {
		return backend.NewBadArgumentError("invalid siteId")
	}

	return self.Backend.DeleteSite(context, siteId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) ListSites(context contextpkg.Context, selectSites backend.SelectSites, window backend.Window) (util.Results[backend.SiteInfo], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	return self.Backend.ListSites(context, selectSites, window)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) PurgeSites(context contextpkg.Context, selectSites backend.SelectSites) error {
	if err := self.Backend.PurgeSites(context, selectSites); err == nil {
		return nil
	} else if backend.IsNotImplementedError(err) {
		if results, err := self.Backend.ListSites(context, selectSites, backend.Window{MaxCount: -1}); err == nil {
			return ParallelDelete(context, results,
				func(siteInfo backend.SiteInfo) string {
					return siteInfo.SiteID
				},
				func(siteId string) error {
					return self.Backend.DeleteSite(context, siteId)
				},
			)
		} else {
			return err
		}
	} else {
		return err
	}
}
