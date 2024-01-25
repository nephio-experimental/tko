package memory

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
)

// ([backend.Backend] interface)
func (self *MemoryBackend) SetPlugin(context contextpkg.Context, plugin *backend.Plugin) error {
	plugin = plugin.Clone()

	if plugin.Properties == nil {
		plugin.Properties = make(map[string]string)
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	self.plugins[plugin.PluginID] = plugin

	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) GetPlugin(context contextpkg.Context, pluginId backend.PluginID) (*backend.Plugin, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if plugin, ok := self.plugins[pluginId]; ok {
		return plugin.Clone(), nil
	} else {
		return nil, backend.NewNotFoundErrorf("plugin: %s", pluginId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeletePlugin(context contextpkg.Context, pluginId backend.PluginID) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if _, ok := self.plugins[pluginId]; ok {
		delete(self.plugins, pluginId)
		return nil
	} else {
		return backend.NewNotFoundErrorf("plugin: %s", pluginId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) ListPlugins(context contextpkg.Context) ([]backend.Plugin, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	plugins := make([]backend.Plugin, len(self.plugins))
	index := 0
	for _, plugin := range self.plugins {
		plugins[index] = *plugin
		index++
	}

	return plugins, nil
}
