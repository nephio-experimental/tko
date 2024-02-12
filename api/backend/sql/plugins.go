package sql

import (
	contextpkg "context"
	"encoding/json"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetPlugin(context contextpkg.Context, plugin *backend.Plugin) error {
	var argumentsJson, propertiesJson []byte
	var err error
	if argumentsJson, err = json.Marshal(plugin.Arguments); err != nil {
		return err
	}
	if propertiesJson, err = json.Marshal(plugin.Properties); err != nil {
		return err
	}

	_, err = self.statements.PreparedInsertPlugin.ExecContext(context, plugin.Type, plugin.Group, plugin.Version, plugin.Kind, plugin.Executor, argumentsJson, propertiesJson)
	return err
}

// ([backend.Backend] interface)
func (self *SQLBackend) GetPlugin(context contextpkg.Context, pluginId backend.PluginID) (*backend.Plugin, error) {
	rows, err := self.statements.PreparedSelectPlugin.QueryContext(context, pluginId.Type, pluginId.Group, pluginId.Version, pluginId.Kind)
	if err != nil {
		return nil, err
	}
	defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

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
func (self *SQLBackend) DeletePlugin(context contextpkg.Context, pluginId backend.PluginID) error {
	if result, err := self.statements.PreparedDeletePlugin.ExecContext(context, pluginId.Type, pluginId.Group, pluginId.Version, pluginId.Kind); err == nil {
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
func (self *SQLBackend) ListPlugins(context contextpkg.Context) (backend.Results[backend.Plugin], error) {
	rows, err := self.statements.PreparedSelectPlugins.QueryContext(context)
	if err != nil {
		return nil, err
	}

	stream := backend.NewResultsStream[backend.Plugin]()

	go func() {
		defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

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
					stream.Close(err)
					return
				}

				if err := jsonUnmarshallMap(propertiesJson, plugin.Properties); err != nil {
					stream.Close(err)
					return
				}

				stream.Send(plugin)
			} else {
				stream.Close(err)
				return
			}
		}

		stream.Close(nil)
	}()

	return stream, nil
}
