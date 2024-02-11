package sql

import (
	contextpkg "context"
	"database/sql"
	"time"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
	"github.com/segmentio/ksuid"
	"github.com/tliron/commonlog"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
	deployment = deployment.Clone()
	if deployment.Metadata == nil {
		deployment.Metadata = make(map[string]string)
	}
	deployment.Update(false)

	if tx, err := self.db.BeginTx(context, nil); err == nil {
		if err := self.mergeDeployment(context, tx, deployment); err != nil {
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return err
		}

		if resources, err := self.encodeResources(deployment.Resources); err == nil {
			insertDeployment := tx.StmtContext(context, self.statements.PreparedInsertDeployment)
			if _, err := insertDeployment.ExecContext(context, deployment.DeploymentID, nilIfEmptyString(deployment.ParentDeploymentID), nilIfEmptyString(deployment.TemplateID), nilIfEmptyString(deployment.SiteID), deployment.Prepared, deployment.Approved, resources); err == nil {
				insertDeploymentMetadata := tx.StmtContext(context, self.statements.PreparedInsertDeploymentMetadata)
				for key, value := range deployment.Metadata {
					if _, err := insertDeploymentMetadata.ExecContext(context, deployment.DeploymentID, key, value); err != nil {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return err
					}
				}

				if deployment.TemplateID != "" {
					insertTemplateDeployment := tx.StmtContext(context, self.statements.PreparedInsertTemplateDeployment)
					if _, err := insertTemplateDeployment.ExecContext(context, deployment.TemplateID, deployment.DeploymentID); err != nil {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return err
					}
				}

				if deployment.SiteID != "" {
					insertSiteDeployment := tx.StmtContext(context, self.statements.PreparedInsertSiteDeployment)
					if _, err := insertSiteDeployment.ExecContext(context, deployment.SiteID, deployment.DeploymentID); err != nil {
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
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return err
		}
	} else {
		return err
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) GetDeployment(context contextpkg.Context, deploymentId string) (*backend.Deployment, error) {
	rows, err := self.statements.PreparedSelectDeployment.QueryContext(context, deploymentId)
	if err != nil {
		return nil, err
	}
	defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

	if rows.Next() {
		var parentDeploymentId, templateId, siteId *string
		var prepared, approved bool
		var metadataJson, resourcesBytes []byte
		if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &metadataJson, &prepared, &approved, &resourcesBytes); err == nil {
			if resources, err := self.decodeResources(resourcesBytes); err == nil {
				deployment := backend.Deployment{
					DeploymentInfo: backend.DeploymentInfo{
						DeploymentID: deploymentId,
						Metadata:     make(map[string]string),
						Prepared:     prepared,
						Approved:     approved,
					},
					Resources: resources,
				}

				if parentDeploymentId != nil {
					deployment.ParentDeploymentID = *parentDeploymentId
				}
				if templateId != nil {
					deployment.TemplateID = *templateId
				}
				if siteId != nil {
					deployment.SiteID = *siteId
				}
				if err := jsonUnmarshallMapEntries(metadataJson, deployment.Metadata); err != nil {
					return nil, err
				}

				return &deployment, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) DeleteDeployment(context contextpkg.Context, deploymentId string) error {
	if result, err := self.statements.PreparedDeleteDeployment.ExecContext(context, deploymentId); err == nil {
		if count, err := result.RowsAffected(); err == nil {
			if count == 0 {
				return backend.NewNotFoundErrorf("deployment: %s", deploymentId)
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
func (self *SQLBackend) ListDeployments(context contextpkg.Context, listDeployments backend.ListDeployments) (backend.DeploymentInfoStream, error) {
	sql := self.statements.SelectDeployments
	var args SqlArgs
	var where SqlWhere
	var with SqlWith

	if (listDeployments.ParentDeploymentID != nil) && (*listDeployments.ParentDeploymentID != "") {
		where.Add("(parent_deployment_id = " + args.Add(listDeployments.ParentDeploymentID) + ")")
	}

	if listDeployments.MetadataPatterns != nil {
		for key, pattern := range listDeployments.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("deployments", "deployment_id", "SELECT deployment_id FROM deployments_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
		}
	}

	for _, pattern := range listDeployments.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("(deployments.template_id ~ " + pattern + ")")
	}

	if listDeployments.TemplateMetadataPatterns != nil {
		for key, pattern := range listDeployments.TemplateMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("deployments", "template_id", "SELECT template_id FROM templates_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
		}
	}

	for _, pattern := range listDeployments.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("(deployments.site_id ~ " + pattern + ")")
	}

	if listDeployments.SiteMetadataPatterns != nil {
		for key, pattern := range listDeployments.SiteMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("deployments", "site_id", "SELECT site_id FROM sites_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
		}
	}

	if listDeployments.Prepared != nil {
		switch *listDeployments.Prepared {
		case true:
			where.Add("prepared")
		case false:
			where.Add("(NOT prepared)")
		}
	}

	if listDeployments.Approved != nil {
		switch *listDeployments.Approved {
		case true:
			where.Add("approved")
		case false:
			where.Add("(NOT approved)")
		}
	}

	sql = where.Apply(sql)
	sql = with.Apply(sql)
	self.log.Infof("generated SQL: %s", sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}
	defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

	var deploymentInfos []backend.DeploymentInfo
	for rows.Next() {
		var deploymentId string
		var parentDeploymentId, templateId, siteId *string
		var metadataJson []byte
		var prepared, approved bool
		if err := rows.Scan(&deploymentId, &parentDeploymentId, &templateId, &siteId, &metadataJson, &prepared, &approved); err == nil {
			deploymentInfo := backend.DeploymentInfo{
				DeploymentID: deploymentId,
				Metadata:     make(map[string]string),
				Prepared:     prepared,
				Approved:     approved,
			}

			if parentDeploymentId != nil {
				deploymentInfo.ParentDeploymentID = *parentDeploymentId
			}
			if templateId != nil {
				deploymentInfo.TemplateID = *templateId
			}
			if siteId != nil {
				deploymentInfo.SiteID = *siteId
			}
			if err := jsonUnmarshallMapEntries(metadataJson, deploymentInfo.Metadata); err != nil {
				return nil, err
			}

			deploymentInfos = append(deploymentInfos, deploymentInfo)
		} else {
			return nil, err
		}
	}

	return backend.NewDeploymentInfoSliceStream(deploymentInfos), nil
}

// ([backend.Backend] interface)
func (self *SQLBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *backend.Deployment, error) {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		selectDeploymentWithModification := tx.StmtContext(context, self.statements.PreparedSelectDeploymentWithModification)
		rows, err := selectDeploymentWithModification.QueryContext(context, deploymentId)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return "", nil, err
		}

		if rows.Next() {
			var parentDeploymentId, templateId, siteId, modificationToken *string
			var prepared, approved bool
			var metadataJson, resources []byte
			var modificationTimestamp *int64
			if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &metadataJson, &prepared, &approved, &resources, &modificationToken, &modificationTimestamp); err == nil {
				if err := rows.Close(); err != nil {
					self.log.Error(err.Error())
				}

				available := (modificationToken == nil) || (*modificationToken == "")
				if !available {
					available = self.hasModificationExpired(modificationTimestamp)
				}

				if available {
					if resources_, err := self.decodeResources(resources); err == nil {
						deployment := backend.Deployment{
							DeploymentInfo: backend.DeploymentInfo{
								DeploymentID: deploymentId,
								Metadata:     make(map[string]string),
								Prepared:     prepared,
								Approved:     approved,
							},
							Resources: resources_,
						}

						if parentDeploymentId != nil {
							deployment.ParentDeploymentID = *parentDeploymentId
						}
						if templateId != nil {
							deployment.TemplateID = *templateId
						}
						if siteId != nil {
							deployment.SiteID = *siteId
						}
						if err := jsonUnmarshallMapEntries(metadataJson, deployment.Metadata); err != nil {
							return "", nil, err
						}

						modificationToken_ := ksuid.New().String()
						modificationTimestamp_ := time.Now().UnixMicro()

						updateDeploymentModification := tx.StmtContext(context, self.statements.PreparedUpdateDeploymentModification)
						if _, err := updateDeploymentModification.ExecContext(context, modificationToken_, modificationTimestamp_, deploymentId); err != nil {
							if err := tx.Rollback(); err != nil {
								self.log.Error(err.Error())
							}
							return "", nil, err
						}

						if err := tx.Commit(); err == nil {
							return modificationToken_, &deployment, nil
						} else {
							return "", nil, err
						}
					} else {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return "", nil, err
					}
				} else {
					if err := tx.Rollback(); err != nil {
						self.log.Error(err.Error())
					}
					return "", nil, backend.NewBusyErrorf("deployment: %s", deploymentId)
				}
			} else {
				if err := rows.Close(); err != nil {
					self.log.Error(err.Error())
				}
				if err := tx.Rollback(); err != nil {
					self.log.Error(err.Error())
				}
				return "", nil, err
			}
		} else {
			if err := rows.Close(); err != nil {
				self.log.Error(err.Error())
			}
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return "", nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
		}
	} else {
		return "", nil, err
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, resources util.Resources) (string, error) {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		selectDeploymentByModification := tx.StmtContext(context, self.statements.PreparedSelectDeploymentByModification)
		rows, err := selectDeploymentByModification.QueryContext(context, modificationToken)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return "", err
		}

		if rows.Next() {
			var deploymentId string
			var parentDeploymentId, templateId, siteId *string
			var prepared, approved bool
			var modificationTimestamp *int64
			if err := rows.Scan(&deploymentId, &parentDeploymentId, &templateId, &siteId, &prepared, &approved, &modificationTimestamp); err == nil {
				if err := rows.Close(); err != nil {
					self.log.Error(err.Error())
				}

				if !self.hasModificationExpired(modificationTimestamp) {
					if resources_, err := self.encodeResources(resources); err == nil {
						deployment := backend.Deployment{
							DeploymentInfo: backend.DeploymentInfo{
								DeploymentID: deploymentId,
								Metadata:     make(map[string]string),
								Prepared:     prepared,
								Approved:     approved,
							},
							Resources: resources,
						}

						deployment.Update(false)

						updateDeployment := tx.StmtContext(context, self.statements.PreparedUpdateDeployment)
						if _, err := updateDeployment.ExecContext(context, nilIfEmptyString(deployment.TemplateID), nilIfEmptyString(deployment.SiteID), deployment.Prepared, deployment.Approved, resources_, deployment.DeploymentID); err != nil {
							if err := tx.Rollback(); err != nil {
								self.log.Error(err.Error())
							}
							return "", err
						}

						// Update metadata

						deleteDeploymentMetadata := tx.StmtContext(context, self.statements.PreparedDeleteDeploymentMetadata)
						if _, err := deleteDeploymentMetadata.ExecContext(context, deploymentId); err != nil {
							if err := tx.Rollback(); err != nil {
								self.log.Error(err.Error())
							}
							return "", err
						}

						insertDeploymentMetadata := tx.StmtContext(context, self.statements.PreparedInsertDeploymentMetadata)
						for key, value := range deployment.Metadata {
							if _, err := insertDeploymentMetadata.ExecContext(context, deploymentId, key, value); err != nil {
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return "", err
							}
						}

						// Update template association

						deleteTemplateDeployments := tx.StmtContext(context, self.statements.PreparedDeleteTemplateDeployments)
						if _, err := deleteTemplateDeployments.ExecContext(context, deploymentId); err != nil {
							if err := tx.Rollback(); err != nil {
								self.log.Error(err.Error())
							}
							return "", err
						}

						if deployment.TemplateID != "" {
							insertTemplateDeployment := tx.StmtContext(context, self.statements.PreparedInsertTemplateDeployment)
							if _, err := insertTemplateDeployment.ExecContext(context, deployment.TemplateID, deploymentId); err != nil {
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return "", err
							}
						}

						// Update site association

						deleteSiteDeployments := tx.StmtContext(context, self.statements.PreparedDeleteSiteDeployments)
						if _, err := deleteSiteDeployments.ExecContext(context, deploymentId); err != nil {
							if err := tx.Rollback(); err != nil {
								self.log.Error(err.Error())
							}
							return "", err
						}

						if deployment.SiteID != "" {
							insertSiteDeployment := tx.StmtContext(context, self.statements.PreparedInsertSiteDeployment)
							if _, err := insertSiteDeployment.ExecContext(context, deployment.SiteID, deploymentId); err != nil {
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return "", err
							}
						}

						if err := tx.Commit(); err == nil {
							return deploymentId, nil
						} else {
							return "", err
						}
					} else {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return "", err
					}
				} else {
					if err := tx.Rollback(); err != nil {
						self.log.Error(err.Error())
					}
					return "", backend.NewNotFoundErrorf("modification token: %s", modificationToken)
				}
			} else {
				if err := rows.Close(); err != nil {
					self.log.Error(err.Error())
				}
				if err := tx.Rollback(); err != nil {
					self.log.Error(err.Error())
				}
				return "", err
			}
		} else {
			if err := rows.Close(); err != nil {
				self.log.Error(err.Error())
			}
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return "", backend.NewNotFoundErrorf("modification token: %s", modificationToken)
		}
	} else {
		return "", err
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) CancelDeploymentModification(context contextpkg.Context, modificationToken string) error {
	if result, err := self.statements.PreparedResetDeploymentModification.ExecContext(context, modificationToken); err == nil {
		if count, err := result.RowsAffected(); err == nil {
			if count == 0 {
				return backend.NewNotFoundErrorf("modification token: %s", modificationToken)
			}
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *SQLBackend) hasModificationExpired(modificationTimestamp *int64) bool {
	if modificationTimestamp == nil {
		return true
	}
	delta := time.Now().UnixMicro() - *modificationTimestamp
	return delta > self.maxModificationDuration
}

func (self *SQLBackend) mergeDeployment(context contextpkg.Context, tx *sql.Tx, deployment *backend.Deployment) error {
	if deployment.TemplateID != "" {
		if template, err := self.getTemplateTx(context, tx, deployment.TemplateID); err == nil {
			deployment.MergeTemplate(template)
		} else {
			return err
		}
	}

	return nil
}
