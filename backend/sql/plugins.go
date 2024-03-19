package sql

import (
	contextpkg "context"
	"database/sql"
	"encoding/json"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
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

	if tx, err := self.db.BeginTx(context, nil); err == nil {
		upsertPlugin := tx.StmtContext(context, self.statements.PreparedUpsertPlugin)
		if _, err := upsertPlugin.ExecContext(context, plugin.Type, plugin.Name, plugin.Executor, argumentsJson, propertiesJson); err == nil {
			if err := self.updatePluginTriggers(context, tx, plugin); err != nil {
				self.rollback(tx)
				return err
			}

			return tx.Commit()
		} else {
			self.rollback(tx)
			return err
		}
	} else {
		return err
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) GetPlugin(context contextpkg.Context, pluginId backend.PluginID) (*backend.Plugin, error) {
	rows, err := self.statements.PreparedSelectPlugin.QueryContext(context, pluginId.Type, pluginId.Name)
	if err != nil {
		return nil, err
	}
	defer self.closeRows(rows)

	if rows.Next() {
		var executor string
		var argumentsJson, propertiesJson, triggersJson []byte
		if err := rows.Scan(&executor, &argumentsJson, &propertiesJson, &triggersJson); err == nil {
			if plugin, err := self.newPlugin(pluginId, executor, argumentsJson, propertiesJson, triggersJson); err == nil {
				return &plugin, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return nil, backend.NewNotFoundErrorf("plugin: %s", pluginId)
}

// ([backend.Backend] interface)
func (self *SQLBackend) DeletePlugin(context contextpkg.Context, pluginId backend.PluginID) error {
	// Will cascade delete plugins_triggers
	if result, err := self.statements.PreparedDeletePlugin.ExecContext(context, pluginId.Type, pluginId.Name); err == nil {
		if count, err := result.RowsAffected(); err == nil {
			if count == 0 {
				return backend.NewNotFoundErrorf("plugin: %s", pluginId)
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
func (self *SQLBackend) ListPlugins(context contextpkg.Context, selectPlugins backend.SelectPlugins, window backend.Window) (util.Results[backend.Plugin], error) {
	sql := self.statements.SelectPlugins
	var with SqlWith
	var where SqlWhere
	var args SqlArgs

	args.AddValue(window.Offset)
	args.AddValue(window.Limit())

	if (selectPlugins.Type != nil) && (*selectPlugins.Type != "") {
		type_ := args.Add(*selectPlugins.Type)
		where.Add(`plugins.type = ` + type_)
	}

	for _, pattern := range selectPlugins.NamePatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`plugins.name ~ ` + pattern)
	}

	if (selectPlugins.Executor != nil) && (*selectPlugins.Executor != "") {
		executor := args.Add(*selectPlugins.Executor)
		where.Add(`plugins.executor = ` + executor)
	}

	if selectPlugins.Trigger != nil {
		group := args.Add(selectPlugins.Trigger.Group)
		version := args.Add(selectPlugins.Trigger.Version)
		kind := args.Add(selectPlugins.Trigger.Kind)
		with.Add(`SELECT plugin_type AS type, plugin_name AS name FROM plugins_triggers WHERE ("group" = `+group+`) AND (version = `+version+`) AND (kind = `+kind+`)`,
			`plugins`, `type`, `name`)
	}

	sql = with.Apply(sql)
	sql = where.Apply(sql)
	self.log.Debugf("generated SQL:\n%s", sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}

	stream := util.NewResultsStream[backend.Plugin](func() {
		self.closeRows(rows)
	})

	go func() {
		for rows.Next() {
			var type_, name, executor string
			var argumentsJson, propertiesJson, triggersJson []byte
			if err := rows.Scan(&type_, &name, &executor, &argumentsJson, &propertiesJson, &triggersJson); err == nil {
				if plugin, err := self.newPlugin(backend.PluginID{Type: type_, Name: name}, executor, argumentsJson, propertiesJson, triggersJson); err == nil {
					stream.Send(plugin)
				} else {
					stream.Close(err)
					return
				}
			} else {
				stream.Close(err)
				return
			}
		}

		stream.Close(nil)
	}()

	return stream, nil
}

// ([backend.Backend] interface)
func (self *SQLBackend) PurgePlugins(context contextpkg.Context, selectPlugins backend.SelectPlugins) error {
	sql := self.statements.DeletePlugins
	var where SqlWhere
	var args SqlArgs

	if (selectPlugins.Type != nil) && (*selectPlugins.Type != "") {
		type_ := args.Add(*selectPlugins.Type)
		where.Add(`plugins.type = ` + type_)
	}

	for _, pattern := range selectPlugins.NamePatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`plugins.name ~ ` + pattern)
	}

	if (selectPlugins.Executor != nil) && (*selectPlugins.Executor != "") {
		executor := args.Add(*selectPlugins.Executor)
		where.Add(`plugins.executor = ` + executor)
	}

	if selectPlugins.Trigger != nil {
		where.Add(`plugins.type = plugins_triggers.plugin_type`)
		where.Add(`plugins.name = plugins_triggers.plugin_name`)
		group := args.Add(selectPlugins.Trigger.Group)
		version := args.Add(selectPlugins.Trigger.Version)
		kind := args.Add(selectPlugins.Trigger.Kind)
		where.Add(`"group" = ` + group)
		where.Add(`version = ` + version)
		where.Add(`kind = ` + kind)
	}

	sql = where.Apply(sql)
	self.log.Debugf("generated SQL:\n%s", sql)

	_, err := self.db.ExecContext(context, sql, args.Args...)
	return err
}

// Utils

func (self *SQLBackend) newPlugin(pluginId backend.PluginID, executor string, argumentsJson []byte, propertiesJson []byte, triggersJson []byte) (backend.Plugin, error) {
	plugin := backend.Plugin{
		PluginID:   pluginId,
		Executor:   executor,
		Properties: make(map[string]string),
	}

	if err := jsonUnmarshallStringArray(argumentsJson, &plugin.Arguments); err != nil {
		return backend.Plugin{}, err
	}

	if err := jsonUnmarshallStringMap(propertiesJson, plugin.Properties); err != nil {
		return backend.Plugin{}, err
	}

	if err := jsonUnmarshallGvkArray(triggersJson, &plugin.Triggers); err != nil {
		return backend.Plugin{}, err
	}

	return plugin, nil
}

func (self *SQLBackend) updatePluginTriggers(context contextpkg.Context, tx *sql.Tx, plugin *backend.Plugin) error {
	deletePluginTriggers := tx.StmtContext(context, self.statements.PreparedDeletePluginTriggers)
	if _, err := deletePluginTriggers.ExecContext(context, plugin.Type, plugin.Name); err != nil {
		return err
	}

	insertPluginTrigger := tx.StmtContext(context, self.statements.PreparedInsertPluginTrigger)
	for _, trigger := range plugin.Triggers {
		if _, err := insertPluginTrigger.ExecContext(context, plugin.Type, plugin.Name, trigger.Group, trigger.Version, trigger.Kind); err != nil {
			return err
		}
	}

	return nil
}
