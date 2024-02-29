package memory

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *MemoryBackend) SetPlugin(context contextpkg.Context, plugin *backend.Plugin) error {
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
func (self *MemoryBackend) ListPlugins(context contextpkg.Context, listPlugins backend.ListPlugins) (util.Results[backend.Plugin], error) {
	self.lock.Lock()

	var plugins []backend.Plugin
	for _, plugin := range self.plugins {
		if (listPlugins.Type != nil) && (*listPlugins.Type != "") {
			if *listPlugins.Type != plugin.Type {
				continue
			}
		}

		if (listPlugins.Executor != nil) && (*listPlugins.Executor != "") {
			if *listPlugins.Executor != plugin.Executor {
				continue
			}
		}

		if !backend.IDMatchesPatterns(plugin.Name, listPlugins.NamePatterns) {
			continue
		}

		if listPlugins.Trigger != nil {
			var found bool
			for _, trigger := range plugin.Triggers {
				if listPlugins.Trigger.Equals(trigger) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		plugins = append(plugins, *plugin)
	}

	self.lock.Unlock()

	backend.SortPlugins(plugins)

	length := uint(len(plugins))
	if listPlugins.Offset > length {
		plugins = nil
	} else if end := listPlugins.Offset + listPlugins.MaxCount; end > length {
		plugins = plugins[listPlugins.Offset:]
	} else {
		plugins = plugins[listPlugins.Offset:end]
	}

	return util.NewResultsSlice(plugins), nil
}
