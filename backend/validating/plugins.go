package validating

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	"github.com/nephio-experimental/tko/plugins"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *ValidatingBackend) SetPlugin(context contextpkg.Context, plugin *backend.Plugin) error {
	if !plugins.IsValidPluginType(plugin.Type, false) {
		return backend.NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, plugin.Type)
	}

	if plugin.Name == "" {
		return backend.NewBadArgumentError("name is empty")
	}
	if !IsValidID(plugin.Name) {
		return backend.NewBadArgumentError("invalid name")
	}

	if plugin.Executor == "" {
		return backend.NewBadArgumentError("executor is empty")
	}

	for _, trigger := range plugin.Triggers {
		// Note: plugin.Group can be empty (for default group)
		if trigger.Version == "" {
			return backend.NewBadArgumentError("trigger ersion is empty")
		}
		if trigger.Kind == "" {
			return backend.NewBadArgumentError("trigger kind is empty")
		}
	}

	return self.Backend.SetPlugin(context, plugin)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) GetPlugin(context contextpkg.Context, pluginId backend.PluginID) (*backend.Plugin, error) {
	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return nil, backend.NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type)
	}

	if pluginId.Name == "" {
		return nil, backend.NewBadArgumentError("name is empty")
	}
	if !IsValidID(pluginId.Name) {
		return nil, backend.NewBadArgumentError("invalid name")
	}

	return self.Backend.GetPlugin(context, pluginId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) DeletePlugin(context contextpkg.Context, pluginId backend.PluginID) error {
	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return backend.NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type)
	}

	if pluginId.Name == "" {
		return backend.NewBadArgumentError("name is empty")
	}
	if !IsValidID(pluginId.Name) {
		return backend.NewBadArgumentError("invalid name")
	}

	return self.Backend.DeletePlugin(context, pluginId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) ListPlugins(context contextpkg.Context, selectPlugins backend.SelectPlugins, window backend.Window) (util.Results[backend.Plugin], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	if selectPlugins.Type != nil {
		if !plugins.IsValidPluginType(*selectPlugins.Type, true) {
			return nil, backend.NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, *selectPlugins.Type)
		}
	}

	if selectPlugins.Trigger != nil {
		// Note: plugin.Group can be empty (for default group)
		if selectPlugins.Trigger.Version == "" {
			return nil, backend.NewBadArgumentError("trigger version is empty")
		}
		if selectPlugins.Trigger.Kind == "" {
			return nil, backend.NewBadArgumentError("trigger kind is empty")
		}
	}

	return self.Backend.ListPlugins(context, selectPlugins, window)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) PurgePlugins(context contextpkg.Context, selectPlugins backend.SelectPlugins) error {
	if err := self.Backend.PurgePlugins(context, selectPlugins); err == nil {
		return nil
	} else if backend.IsNotImplementedError(err) {
		if results, err := self.Backend.ListPlugins(context, selectPlugins, backend.Window{MaxCount: -1}); err == nil {
			return ParallelDelete(context, results,
				func(plugin backend.Plugin) backend.PluginID {
					return plugin.PluginID
				},
				func(pluginId backend.PluginID) error {
					return self.Backend.DeletePlugin(context, pluginId)
				},
			)
		} else {
			return err
		}
	} else {
		return err
	}
}
