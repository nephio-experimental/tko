package sql

import (
	"encoding/json"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
)

// ([backend.Backend] interface)
func (self *SqlBackend) SetPlugin(plugin *backend.Plugin) error {
	var argumentsJson, propertiesJson []byte
	var err error
	if argumentsJson, err = json.Marshal(plugin.Arguments); err != nil {
		return err
	}
	if propertiesJson, err = json.Marshal(plugin.Properties); err != nil {
		return err
	}

	_, err = self.sql.PreparedInsertPlugin.Exec(plugin.Type, plugin.Group, plugin.Version, plugin.Kind, plugin.Executor, argumentsJson, propertiesJson)
	return err
}

// ([backend.Backend] interface)
func (self *SqlBackend) GetPlugin(pluginId backend.PluginID) (*backend.Plugin, error) {
	rows, err := self.sql.PreparedSelectPlugin.Query(pluginId.Type, pluginId.Group, pluginId.Version, pluginId.Kind)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			self.log.Error(err.Error())
		}
	}()

	if rows.Next() {
		var executor string
		var argumentsJson, propertiesJson []byte
		if err := rows.Scan(&executor, &argumentsJson, &propertiesJson); err == nil {
			plugin := backend.Plugin{
				PluginID:   pluginId,
				Executor:   executor,
				Properties: make(map[string]string),
			}

			if err := jsonUnmarshallArray(argumentsJson, &plugin.Arguments); err != nil {
				return nil, err
			}

			if err := jsonUnmarshallMap(propertiesJson, plugin.Properties); err != nil {
				return nil, err
			}

			return &plugin, nil
		} else {
			return nil, err
		}
	} else {
		return nil, backend.NewNotFoundErrorf("plugin: %s", pluginId)
	}
}

// ([backend.Backend] interface)
func (self *SqlBackend) DeletePlugin(pluginId backend.PluginID) error {
	if result, err := self.sql.PreparedDeletePlugin.Exec(pluginId.Type, pluginId.Group, pluginId.Version, pluginId.Kind); err == nil {
		if count, err := result.RowsAffected(); err == nil {
			if count == 0 {
				return backend.NewNotFoundErrorf("deployment: %s", pluginId)
			}
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

// ([backend.Backend] interface)
func (self *SqlBackend) ListPlugins() ([]backend.Plugin, error) {
	rows, err := self.sql.PreparedSelectPlugins.Query()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			self.log.Error(err.Error())
		}
	}()

	var plugins []backend.Plugin
	for rows.Next() {
		var type_, group, version, kind, executor string
		var argumentsJson, propertiesJson []byte
		if err := rows.Scan(&type_, &group, &version, &kind, &executor, &argumentsJson, &propertiesJson); err == nil {
			plugin := backend.Plugin{
				PluginID: backend.PluginID{
					Type: type_,
					GVK: util.GVK{
						Group:   group,
						Version: version,
						Kind:    kind,
					},
				},
				Executor:   executor,
				Properties: make(map[string]string),
			}

			if err := jsonUnmarshallArray(argumentsJson, &plugin.Arguments); err != nil {
				return nil, err
			}

			if err := jsonUnmarshallMap(propertiesJson, plugin.Properties); err != nil {
				return nil, err
			}

			plugins = append(plugins, plugin)
		} else {
			return nil, err
		}
	}

	return plugins, nil
}
