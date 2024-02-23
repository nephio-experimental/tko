package sql

import (
	contextpkg "context"
	"database/sql"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetTemplate(context contextpkg.Context, template *backend.Template) error {
	if resources, err := self.encodeResources(template.Resources); err == nil {
		if tx, err := self.db.BeginTx(context, nil); err == nil {
			upsertTemplate := tx.StmtContext(context, self.statements.PreparedUpsertTemplate)
			if _, err := upsertTemplate.ExecContext(context, template.TemplateID, resources); err == nil {
				if err := self.updateTemplateMetadata(context, tx, template); err != nil {
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
	} else {
		return err
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) GetTemplate(context contextpkg.Context, templateId string) (*backend.Template, error) {
	return self.getTemplateStmt(context, self.statements.PreparedSelectTemplate, templateId)
}

// ([backend.Backend] interface)
func (self *SQLBackend) DeleteTemplate(context contextpkg.Context, templateId string) error {
	// Will cascade delete templates_metadata, templates_deployments
	if result, err := self.statements.PreparedDeleteTemplate.ExecContext(context, templateId); err == nil {
		if count, err := result.RowsAffected(); err == nil {
			if count == 0 {
				return backend.NewNotFoundErrorf("template: %s", templateId)
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
func (self *SQLBackend) ListTemplates(context contextpkg.Context, listTemplates backend.ListTemplates) (util.Results[backend.TemplateInfo], error) {
	sql := self.statements.SelectTemplates
	var with SqlWith
	var where SqlWhere
	var args SqlArgs

	for _, pattern := range listTemplates.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("templates.template_id ~ " + pattern)
	}

	if listTemplates.MetadataPatterns != nil {
		for key, pattern := range listTemplates.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("SELECT template_id FROM templates_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")",
				"templates", "template_id")
		}
	}

	sql = with.Apply(sql)
	sql = where.Apply(sql)
	self.log.Debugf("generated SQL:\n%s", sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}

	stream := util.NewResultsStream[backend.TemplateInfo](func() {
		self.closeRows(rows)
	})

	go func() {
		for rows.Next() {
			var templateId string
			var metadataJson, deploymentIdsJson []byte
			if err := rows.Scan(&templateId, &metadataJson, &deploymentIdsJson); err == nil {
				if templateInfo, err := self.newTemplateInfo(templateId, metadataJson, deploymentIdsJson); err == nil {
					stream.Send(templateInfo)
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

// Utils

func (self *SQLBackend) newTemplateInfo(templateId string, metadataJson []byte, deploymentIdsJson []byte) (backend.TemplateInfo, error) {
	templateInfo := backend.TemplateInfo{
		TemplateID: templateId,
		Metadata:   make(map[string]string),
	}

	if err := jsonUnmarshallStringMapEntries(metadataJson, templateInfo.Metadata); err != nil {
		return backend.TemplateInfo{}, err
	}

	if err := jsonUnmarshallStringArray(deploymentIdsJson, &templateInfo.DeploymentIDs); err != nil {
		return backend.TemplateInfo{}, err
	}

	return templateInfo, nil
}

func (self *SQLBackend) newTemplate(templateId string, metadataJson []byte, deploymentIdsJson []byte, resources []byte) (*backend.Template, error) {
	if templateInfo, err := self.newTemplateInfo(templateId, metadataJson, deploymentIdsJson); err == nil {
		template := backend.Template{TemplateInfo: templateInfo}
		if template.Resources, err = self.decodeResources(resources); err == nil {
			return &template, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *SQLBackend) getTemplateTx(context contextpkg.Context, tx *sql.Tx, templateId string) (*backend.Template, error) {
	selectTemplate := tx.StmtContext(context, self.statements.PreparedSelectTemplate)
	return self.getTemplateStmt(context, selectTemplate, templateId)
}

func (self *SQLBackend) getTemplateStmt(context contextpkg.Context, selectTemplate *sql.Stmt, templateId string) (*backend.Template, error) {
	rows, err := selectTemplate.QueryContext(context, templateId)
	if err != nil {
		return nil, err
	}
	defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

	if rows.Next() {
		var resources, metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&resources, &metadataJson, &deploymentIdsJson); err == nil {
			return self.newTemplate(templateId, metadataJson, deploymentIdsJson, resources)
		} else {
			return nil, err
		}
	}

	return nil, backend.NewNotFoundErrorf("template: %s", templateId)
}

func (self *SQLBackend) updateTemplateMetadata(context contextpkg.Context, tx *sql.Tx, template *backend.Template) error {
	deleteTemplateMetadata := tx.StmtContext(context, self.statements.PreparedDeleteTemplateMetadata)
	if _, err := deleteTemplateMetadata.ExecContext(context, template.TemplateID); err != nil {
		return err
	}

	upsertTemplateMetadata := tx.StmtContext(context, self.statements.PreparedUpsertTemplateMetadata)
	for key, value := range template.Metadata {
		if _, err := upsertTemplateMetadata.ExecContext(context, template.TemplateID, key, value); err != nil {
			return err
		}
	}

	return nil
}
