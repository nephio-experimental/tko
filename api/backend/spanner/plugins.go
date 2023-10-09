package spanner

import (
	"github.com/nephio-experimental/tko/api/backend"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetPlugin(plugin *backend.Plugin) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetPlugin(pluginId backend.PluginID) (*backend.Plugin, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeletePlugin(pluginId backend.PluginID) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListPlugins() ([]backend.Plugin, error) {
	return nil, nil
}
