package sql

import (
	contextpkg "context"
	"database/sql"
	"time"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetTemplate(context contextpkg.Context, template *backend.Template) error {
	var package_ []byte
	var err error
	if package_, err = self.encodePackage(template.Package); err != nil {
		return err
	}

	if tx, err := self.db.BeginTx(context, nil); err == nil {
		template.Updated = time.Now().UTC()
		upsertTemplate := tx.StmtContext(context, self.statements.PreparedUpsertTemplate)
		if _, err := upsertTemplate.ExecContext(context, template.TemplateID, template.Updated, package_); err == nil {
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
func (self *SQLBackend) ListTemplates(context contextpkg.Context, selectTemplates backend.SelectTemplates, window backend.Window) (util.Results[backend.TemplateInfo], error) {
	sql := self.statements.SelectTemplates
	var args SqlArgs
	var with SqlWith
	var where SqlWhere

	args.AddValue(window.Offset)
	args.AddValue(window.Limit())

	for _, pattern := range selectTemplates.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`templates.template_id ~ ` + pattern)
	}

	if len(selectTemplates.MetadataPatterns) > 0 {
		for key, pattern := range selectTemplates.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add(`SELECT template_id FROM templates_metadata WHERE (key = `+key+`) AND (value ~ `+pattern+`)`,
				`templates`, `template_id`)
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
			var updated time.Time
			var metadataJson, deploymentIdsJson []byte
			if err := rows.Scan(&templateId, &updated, &metadataJson, &deploymentIdsJson); err == nil {
				if templateInfo, err := self.newTemplateInfo(templateId, updated, metadataJson, deploymentIdsJson); err == nil {
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

// ([backend.Backend] interface)
func (self *SQLBackend) PurgeTemplates(context contextpkg.Context, selectTemplates backend.SelectTemplates) error {
	sql := self.statements.DeleteTemplates
	var args SqlArgs
	var where SqlWhere

	for _, pattern := range selectTemplates.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`template_id ~ ` + pattern)
	}

	if len(selectTemplates.MetadataPatterns) > 0 {
		where.Add(`templates.template_id = templates_metadata.template_id`)
		for key, pattern := range selectTemplates.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			where.Add(`key = ` + key)
			where.Add(`value ~ ` + pattern)
		}
	}

	sql = where.Apply(sql)
	self.log.Debugf("generated SQL:\n%s", sql)

	_, err := self.db.ExecContext(context, sql, args.Args...)
	return err
}

// Utils

func (self *SQLBackend) newTemplateInfo(templateId string, updated time.Time, metadataJson []byte, deploymentIdsJson []byte) (backend.TemplateInfo, error) {
	templateInfo := backend.TemplateInfo{
		TemplateID: templateId,
		Metadata:   make(map[string]string),
		Updated:    updated,
	}

	if err := jsonUnmarshallStringMapEntries(metadataJson, templateInfo.Metadata); err != nil {
		return backend.TemplateInfo{}, err
	}

	if err := jsonUnmarshallStringArray(deploymentIdsJson, &templateInfo.DeploymentIDs); err != nil {
		return backend.TemplateInfo{}, err
	}

	return templateInfo, nil
}

func (self *SQLBackend) newTemplate(templateId string, updated time.Time, metadataJson []byte, deploymentIdsJson []byte, package_ []byte) (*backend.Template, error) {
	if templateInfo, err := self.newTemplateInfo(templateId, updated, metadataJson, deploymentIdsJson); err == nil {
		template := backend.Template{TemplateInfo: templateInfo}
		if template.Package, err = self.decodePackage(package_); err == nil {
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
	defer self.closeRows(rows)

	if rows.Next() {
		var updated time.Time
		var package_, metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&updated, &package_, &metadataJson, &deploymentIdsJson); err == nil {
			return self.newTemplate(templateId, updated, metadataJson, deploymentIdsJson, package_)
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
