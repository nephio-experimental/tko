package spanner

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
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
func (self *SpannerBackend) ListPlugins(context contextpkg.Context, listPlugins backend.ListPlugins) (util.Results[backend.Plugin], error) {
	return nil, nil
}
