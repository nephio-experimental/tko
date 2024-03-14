package spanner

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetPlugin(context contextpkg.Context, plugin *backend.Plugin) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetPlugin(context contextpkg.Context, pluginId backend.PluginID) (*backend.Plugin, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeletePlugin(context contextpkg.Context, pluginId backend.PluginID) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListPlugins(context contextpkg.Context, selectPlugins backend.SelectPlugins, window backend.Window) (util.Results[backend.Plugin], error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) PurgePlugins(context contextpkg.Context, selectPlugins backend.SelectPlugins) error {
	return backend.NewNotImplementedError("PurgePlugins")
}
