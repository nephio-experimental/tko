package sql

import (
	contextpkg "context"
	"time"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
	"github.com/segmentio/ksuid"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
	deployment = deployment.Clone()
	deployment.UpdateInfo(false)

	if tx, err := self.db.BeginTx(context, nil); err == nil {
		if deployment.TemplateID != "" {
			if resources, err := self.getTemplateResources(context, tx, deployment.TemplateID); err == nil {
				// Merge our resources over template resources
				resources = util.MergeResources(resources, deployment.Resources)

				// Merge default Deployment resource
				resources = util.MergeResources(resources, util.Resources{util.NewDeploymentResource(deployment.TemplateID, deployment.SiteID, deployment.Prepared)})

				deployment.Resources = resources
			} else {
				return err
			}
		}

		if resources, err := self.encodeResources(deployment.Resources); err == nil {
			if _, err := tx.ExecContext(context, self.statements.InsertDeployment, deployment.DeploymentID, nilIfEmptyString(deployment.ParentDeploymentID), nilIfEmptyString(deployment.TemplateID), nilIfEmptyString(deployment.SiteID), deployment.Prepared, resources); err == nil {
				if deployment.TemplateID != "" {
					if _, err := tx.ExecContext(context, self.statements.InsertTemplateDeployment, deployment.TemplateID, deployment.DeploymentID); err != nil {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return err
					}
				}

				if deployment.SiteID != "" {
					if _, err := tx.ExecContext(context, self.statements.InsertSiteDeployment, deployment.SiteID, deployment.DeploymentID); err != nil {
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
	defer func() {
		if err := rows.Close(); err != nil {
			self.log.Error(err.Error())
		}
	}()

	if rows.Next() {
		var parentDeploymentId, templateId, siteId *string
		var prepared bool
		var resources []byte
		if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &prepared, &resources); err == nil {
			if resources_, err := self.decodeResources(resources); err == nil {
				deployment := backend.Deployment{
					DeploymentInfo: backend.DeploymentInfo{
						DeploymentID: deploymentId,
						Prepared:     prepared,
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
func (self *SQLBackend) ListDeployments(context contextpkg.Context, prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]backend.DeploymentInfo, error) {
	sql := self.statements.SelectDeployments
	var args SqlArgs
	var where SqlWhere
	var with SqlWith

	switch prepared {
	case "true":
		where.Add("prepared")
	case "false":
		where.Add("(NOT prepared)")
	}

	if parentDeploymentId != "" {
		where.Add("(parent_deployment_id = " + args.Add(parentDeploymentId) + ")")
	}

	for _, pattern := range templateIdPatterns {
		pattern = args.Add(backend.IdPatternRE(pattern))
		where.Add("(deployments.template_id ~ " + pattern + ")")
	}

	if templateMetadataPatterns != nil {
		for key, pattern := range templateMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("deployments", "template_id", "SELECT template_id FROM templates_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
		}
	}

	for _, pattern := range siteIdPatterns {
		pattern = args.Add(backend.IdPatternRE(pattern))
		where.Add("(deployments.site_id ~ " + pattern + ")")
	}

	if siteMetadataPatterns != nil {
		for key, pattern := range siteMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("deployments", "site_id", "SELECT site_id FROM sites_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
		}
	}

	sql = where.Apply(sql)
	sql = with.Apply(sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			self.log.Error(err.Error())
		}
	}()

	var deploymentInfos []backend.DeploymentInfo
	for rows.Next() {
		var deploymentId string
		var parentDeploymentId, templateId, siteId *string
		var prepared bool
		if err := rows.Scan(&deploymentId, &parentDeploymentId, &templateId, &siteId, &prepared); err == nil {
			deploymentInfo := backend.DeploymentInfo{
				DeploymentID: deploymentId,
				Prepared:     prepared,
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

			deploymentInfos = append(deploymentInfos, deploymentInfo)
		} else {
			return nil, err
		}
	}

	return deploymentInfos, nil
}

// ([backend.Backend] interface)
func (self *SQLBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *backend.Deployment, error) {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		rows, err := tx.QueryContext(context, self.statements.SelectDeploymentWithModificaiton, deploymentId)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return "", nil, err
		}

		if rows.Next() {
			var parentDeploymentId, templateId, siteId, modificationToken *string
			var prepared bool
			var resources []byte
			var modificationTimestamp *int64
			if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &prepared, &resources, &modificationToken, &modificationTimestamp); err == nil {
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
								Prepared:     prepared,
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

						modificationToken_ := ksuid.New().String()
						modificationTimestamp_ := time.Now().UnixMicro()

						if _, err := tx.ExecContext(context, self.statements.UpdateDeploymentModification, modificationToken_, modificationTimestamp_, deploymentId); err != nil {
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
		rows, err := tx.QueryContext(context, self.statements.SelectDeploymentByModification, modificationToken)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return "", err
		}

		if rows.Next() {
			var deploymentId string
			var parentDeploymentId, templateId, siteId *string
			var prepared bool
			var modificationTimestamp *int64
			if err := rows.Scan(&deploymentId, &parentDeploymentId, &templateId, &siteId, &prepared, &modificationTimestamp); err == nil {
				if err := rows.Close(); err != nil {
					self.log.Error(err.Error())
				}

				if !self.hasModificationExpired(modificationTimestamp) {
					if resources_, err := self.encodeResources(resources); err == nil {
						deployment := backend.Deployment{
							DeploymentInfo: backend.DeploymentInfo{
								DeploymentID: deploymentId,
							},
							Resources: resources,
						}

						deployment.UpdateInfo(false)

						if _, err := tx.ExecContext(context, self.statements.UpdateDeployment, nilIfEmptyString(deployment.TemplateID), nilIfEmptyString(deployment.SiteID), deployment.Prepared, resources_, deployment.DeploymentID); err != nil {
							return "", err
						}

						if deployment.TemplateID != "" {
							if _, err := tx.ExecContext(context, self.statements.InsertTemplateDeployment, deployment.TemplateID, deploymentId); err != nil {
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return "", err
							}
						}

						if deployment.SiteID != "" {
							if _, err := tx.ExecContext(context, self.statements.InsertSiteDeployment, deployment.SiteID, deploymentId); err != nil {
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
