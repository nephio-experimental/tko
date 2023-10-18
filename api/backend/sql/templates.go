package sql

import (
	"database/sql"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
)

// ([backend.Backend] interface)
func (self *SqlBackend) SetTemplate(template *backend.Template) error {
	template = template.Clone()
	if template.Metadata == nil {
		template.Metadata = make(map[string]string)
	}

	template.Update()

	if resources, err := self.encodeResources(template.Resources); err == nil {
		if tx, err := self.db.Begin(); err == nil {
			if _, err := tx.Exec(self.sql.InsertTemplate, template.TemplateID, resources); err == nil {
				if len(template.Metadata) > 0 {
					if insertTemplateMetadata, err := tx.Prepare(self.sql.InsertTemplateMetadata); err == nil {
						for key, value := range template.Metadata {
							if _, err := insertTemplateMetadata.Exec(template.TemplateID, key, value); err != nil {
								insertTemplateMetadata.Close()
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return err
							}
						}
						insertTemplateMetadata.Close()
					} else {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return err
					}
				}

				if len(template.DeploymentIDs) > 0 {
					if insertTemplateDeployment, err := tx.Prepare(self.sql.InsertTemplateDeployment); err == nil {
						for _, deploymentId := range template.DeploymentIDs {
							if _, err := insertTemplateDeployment.Exec(template.TemplateID, deploymentId); err != nil {
								insertTemplateDeployment.Close()
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return err
							}
						}
						insertTemplateDeployment.Close()
					} else {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return err
					}
				}

				return tx.Commit()
			} else {
				if err := tx.Rollback(); err != nil {
					self.log.Error(err.Error())
				}
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
func (self *SqlBackend) GetTemplate(templateId string) (*backend.Template, error) {
	rows, err := self.sql.PreparedSelectTemplate.Query(templateId)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			self.log.Error(err.Error())
		}
	}()

	if rows.Next() {
		var resources []byte
		var metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&resources, &metadataJson, &deploymentIdsJson); err == nil {
			template := backend.Template{
				TemplateInfo: backend.TemplateInfo{
					TemplateID: templateId,
					Metadata:   make(map[string]string),
				},
			}

			if template.Resources == nil {
				if template.Resources, err = self.decodeResources(resources); err != nil {
					return nil, err
				}
			}

			if err := jsonUnmarshallMapEntries(metadataJson, template.Metadata); err != nil {
				return nil, err
			}

			if err := jsonUnmarshallArray(deploymentIdsJson, &template.DeploymentIDs); err != nil {
				return nil, err
			}

			return &template, nil
		} else {
			return nil, err
		}
	}

	return nil, backend.NewNotFoundErrorf("template: %s", templateId)
}

// ([backend.Backend] interface)
func (self *SqlBackend) DeleteTemplate(templateId string) error {
	if result, err := self.sql.PreparedDeleteTemplate.Exec(templateId); err == nil {
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
func (self *SqlBackend) ListTemplates(templateIdPatterns []string, metadataPatterns map[string]string) ([]backend.TemplateInfo, error) {
	sql := self.sql.SelectTemplates
	var args SqlArgs
	var where SqlWhere
	var with SqlWith

	for _, pattern := range templateIdPatterns {
		pattern = args.Add(backend.IdPatternRE(pattern))
		where.Add("(templates.template_id ~ " + pattern + ")")
	}

	if metadataPatterns != nil {
		for key, pattern := range metadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("templates", "template_id", "SELECT template_id FROM templates_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
		}
	}

	sql = where.Apply(sql)
	sql = with.Apply(sql)

	rows, err := self.db.Query(sql, args.Args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			self.log.Error(err.Error())
		}
	}()

	var templateInfos []backend.TemplateInfo
	for rows.Next() {
		var templateId string
		var metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&templateId, &metadataJson, &deploymentIdsJson); err == nil {
			templateInfo := backend.TemplateInfo{
				TemplateID: templateId,
				Metadata:   make(map[string]string),
			}

			if err := jsonUnmarshallMapEntries(metadataJson, templateInfo.Metadata); err != nil {
				return nil, err
			}

			if err := jsonUnmarshallArray(deploymentIdsJson, &templateInfo.DeploymentIDs); err != nil {
				return nil, err
			}

			templateInfos = append(templateInfos, templateInfo)
		} else {
			return nil, err
		}
	}

	return templateInfos, nil
}

// Utils

func (self *SqlBackend) getTemplateResources(tx *sql.Tx, templateId string) (util.Resources, error) {
	if rows, err := tx.Query(self.sql.SelectTemplateResources, templateId); err == nil {
		defer func() {
			if err := rows.Close(); err != nil {
				self.log.Error(err.Error())
			}
		}()

		if rows.Next() {
			var resources []byte
			if err := rows.Scan(&resources); err == nil {
				return self.decodeResources(resources)
			} else {
				return nil, err
			}
		}

		return nil, backend.NewNotFoundErrorf("template: %s", templateId)
	} else {
		return nil, err
	}
}
