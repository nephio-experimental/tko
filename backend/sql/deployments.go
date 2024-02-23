package sql

import (
	contextpkg "context"
	"database/sql"
	"time"

	"github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SQLBackend) CreateDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		if err := self.mergeDeploymentTemplate(context, tx, deployment); err != nil {
			self.rollback(tx)
			return err
		}
		deployment.MergeDeploymentResource()
		deployment.UpdateFromResources(true)

		var resources []byte
		var err error
		if resources, err = self.encodeResources(deployment.Resources); err != nil {
			self.rollback(tx)
			return err
		}

		insertDeployment := tx.StmtContext(context, self.statements.PreparedInsertDeployment)
		if _, err := insertDeployment.ExecContext(context, deployment.DeploymentID, nilIfEmptyString(deployment.ParentDeploymentID), nilIfEmptyString(deployment.TemplateID), nilIfEmptyString(deployment.SiteID), deployment.Prepared, deployment.Approved, resources); err != nil {
			self.rollback(tx)
			return err
		}

		if err := self.upsertDeploymentMetadata(context, tx, deployment); err != nil {
			self.rollback(tx)
			return err
		}

		if err := self.upsertTemplateDeployment(context, tx, deployment); err != nil {
			self.rollback(tx)
			return err
		}

		if err := self.upsertSiteDeployment(context, tx, deployment); err != nil {
			self.rollback(tx)
			return err
		}

		return tx.Commit()
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
		var metadataJson, resources []byte
		if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &metadataJson, &prepared, &approved, &resources); err == nil {
			return self.newDeployment(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, prepared, approved, resources)
		} else {
			return nil, err
		}
	}

	return nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
}

// ([backend.Backend] interface)
func (self *SQLBackend) DeleteDeployment(context contextpkg.Context, deploymentId string) error {
	// Will cascade delete deployments_metadata, templates_deployments, sites_deployments
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
func (self *SQLBackend) ListDeployments(context contextpkg.Context, listDeployments backend.ListDeployments) (util.Results[backend.DeploymentInfo], error) {
	sql := self.statements.SelectDeployments
	var with SqlWith
	var where SqlWhere
	var args SqlArgs

	if (listDeployments.ParentDeploymentID != nil) && (*listDeployments.ParentDeploymentID != "") {
		where.Add("parent_deployment_id = " + args.Add(listDeployments.ParentDeploymentID))
	}

	if listDeployments.MetadataPatterns != nil {
		for key, pattern := range listDeployments.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("SELECT deployment_id FROM deployments_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")",
				"deployments", "deployment_id")
		}
	}

	for _, pattern := range listDeployments.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("deployments.template_id ~ " + pattern)
	}

	if listDeployments.TemplateMetadataPatterns != nil {
		for key, pattern := range listDeployments.TemplateMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("SELECT template_id FROM templates_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")",
				"deployments", "template_id")
		}
	}

	for _, pattern := range listDeployments.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("deployments.site_id ~ " + pattern)
	}

	if listDeployments.SiteMetadataPatterns != nil {
		for key, pattern := range listDeployments.SiteMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("SELECT site_id FROM sites_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")",
				"deployments", "site_id")
		}
	}

	if listDeployments.Prepared != nil {
		switch *listDeployments.Prepared {
		case true:
			where.Add("prepared")
		case false:
			where.Add("NOT prepared")
		}
	}

	if listDeployments.Approved != nil {
		switch *listDeployments.Approved {
		case true:
			where.Add("approved")
		case false:
			where.Add("NOT approved")
		}
	}

	sql = with.Apply(sql)
	sql = where.Apply(sql)
	self.log.Infof("generated SQL:\n%s", sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}

	stream := util.NewResultsStream[backend.DeploymentInfo](func() {
		self.closeRows(rows)
	})

	go func() {
		defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

		for rows.Next() {
			var deploymentId string
			var parentDeploymentId, templateId, siteId *string
			var metadataJson []byte
			var prepared, approved bool
			if err := rows.Scan(&deploymentId, &parentDeploymentId, &templateId, &siteId, &metadataJson, &prepared, &approved); err == nil {
				if deploymentInfo, err := self.newDeploymentInfo(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, prepared, approved); err == nil {
					stream.Send(deploymentInfo)
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
func (self *SQLBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *backend.Deployment, error) {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		selectDeploymentWithModification := tx.StmtContext(context, self.statements.PreparedSelectDeploymentWithModification)
		rows, err := selectDeploymentWithModification.QueryContext(context, deploymentId)
		if err != nil {
			self.rollback(tx)
			return "", nil, err
		}

		if rows.Next() {
			var parentDeploymentId, templateId, siteId, modificationToken *string
			var prepared, approved bool
			var metadataJson, resources []byte
			var modificationTimestamp *int64
			if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &metadataJson, &prepared, &approved, &resources, &modificationToken, &modificationTimestamp); err == nil {
				self.closeRows(rows)

				available := (modificationToken == nil) || (*modificationToken == "")
				if !available {
					available = self.hasModificationExpired(modificationTimestamp)
				}

				if !available {
					self.rollback(tx)
					return "", nil, backend.NewBusyErrorf("deployment: %s", deploymentId)
				}

				if deployment, err := self.newDeployment(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, prepared, approved, resources); err == nil {
					modificationToken_ := backend.NewID()
					modificationTimestamp_ := time.Now().UnixMicro()

					updateDeploymentModification := tx.StmtContext(context, self.statements.PreparedUpdateDeploymentModification)
					if _, err := updateDeploymentModification.ExecContext(context, modificationToken_, modificationTimestamp_, deploymentId); err != nil {
						self.rollback(tx)
						return "", nil, err
					}

					if err := tx.Commit(); err == nil {
						return modificationToken_, deployment, nil
					} else {
						return "", nil, err
					}
				} else {
					self.rollback(tx)
					return "", nil, err
				}
			} else {
				self.closeRows(rows)
				self.rollback(tx)
				return "", nil, err
			}
		} else {
			self.closeRows(rows)
			self.rollback(tx)
			return "", nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
		}
	} else {
		return "", nil, err
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, resources tkoutil.Resources, validation *validationpkg.Validation) (string, error) {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		selectDeploymentByModification := tx.StmtContext(context, self.statements.PreparedSelectDeploymentByModification)
		rows, err := selectDeploymentByModification.QueryContext(context, modificationToken)
		if err != nil {
			self.rollback(tx)
			return "", err
		}

		if rows.Next() {
			var deploymentId string
			var templateId, siteId *string
			var metadataJson []byte
			var prepared, approved bool
			var modificationTimestamp *int64
			if err := rows.Scan(&deploymentId, &templateId, &siteId, &metadataJson, &prepared, &approved, &modificationTimestamp); err == nil {
				self.closeRows(rows)

				if self.hasModificationExpired(modificationTimestamp) {
					self.rollback(tx)
					return "", backend.NewNotFoundErrorf("modification token: %s", modificationToken)
				}

				var deploymentInfo backend.DeploymentInfo
				if deploymentInfo, err = self.newDeploymentInfo(deploymentId, nil, templateId, siteId, metadataJson, prepared, approved); err != nil {
					self.rollback(tx)
					return "", err
				}

				deployment := backend.Deployment{
					DeploymentInfo: deploymentInfo,
					Resources:      resources,
				}

				originalTemplateId := deploymentInfo.TemplateID
				originalSiteId := deploymentInfo.SiteID
				originalMetadata := tkoutil.CloneStringMap(deployment.Metadata)
				deployment.UpdateFromResources(false)

				if validation != nil {
					// Complete validation when fully prepared
					if err := validation.ValidateResources(resources, deployment.Prepared); err != nil {
						self.rollback(tx)
						return "", err
					}
				}

				var resources []byte
				var err error
				if resources, err = self.encodeResources(deployment.Resources); err != nil {
					self.rollback(tx)
					return "", err
				}

				updateDeployment := tx.StmtContext(context, self.statements.PreparedUpdateDeployment)
				if _, err := updateDeployment.ExecContext(context, deployment.DeploymentID, deployment.Prepared, deployment.Approved, resources); err != nil {
					self.rollback(tx)
					return "", err
				}

				// Update metadata

				if !tkoutil.StringMapEquals(originalMetadata, deployment.Metadata) {
					if err := self.updateDeploymentMetadata(context, tx, &deployment); err != nil {
						self.rollback(tx)
						return "", err
					}
				}

				// Update template association

				if deployment.TemplateID != originalTemplateId {
					if deployment.TemplateID != "" {
						if err := self.upsertTemplateDeployment(context, tx, &deployment); err != nil {
							self.rollback(tx)
							return "", err
						}
					} else {
						deleteTemplateDeployment := tx.StmtContext(context, self.statements.PreparedDeleteTemplateDeployment)
						if _, err := deleteTemplateDeployment.ExecContext(context, deployment.DeploymentID); err != nil {
							self.rollback(tx)
							return "", err
						}
					}
				}

				// Update site association

				if deployment.SiteID != originalSiteId {
					if deployment.SiteID != "" {
						if err := self.upsertSiteDeployment(context, tx, &deployment); err != nil {
							self.rollback(tx)
							return "", err
						}
					} else {
						deleteSiteDeployment := tx.StmtContext(context, self.statements.PreparedDeleteSiteDeployment)
						if _, err := deleteSiteDeployment.ExecContext(context, deployment.DeploymentID); err != nil {
							self.rollback(tx)
							return "", err
						}
					}
				}

				if err := tx.Commit(); err == nil {
					return deploymentId, nil
				} else {
					return "", err
				}
			} else {
				self.closeRows(rows)
				self.rollback(tx)
				return "", err
			}
		} else {
			self.closeRows(rows)
			self.rollback(tx)
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

// Utils

func (self *SQLBackend) newDeploymentInfo(deploymentId string, parentDeploymentId *string, templateId *string, siteId *string, metadataJson []byte, prepared bool, approved bool) (backend.DeploymentInfo, error) {
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

	if err := jsonUnmarshallStringMapEntries(metadataJson, deploymentInfo.Metadata); err != nil {
		return backend.DeploymentInfo{}, err
	}

	return deploymentInfo, nil
}

func (self *SQLBackend) newDeployment(deploymentId string, parentDeploymentId *string, templateId *string, siteId *string, metadataJson []byte, prepared bool, approved bool, resources []byte) (*backend.Deployment, error) {
	if deploymentInfo, err := self.newDeploymentInfo(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, prepared, approved); err == nil {
		deployment := backend.Deployment{DeploymentInfo: deploymentInfo}
		if deployment.Resources, err = self.decodeResources(resources); err == nil {
			return &deployment, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *SQLBackend) mergeDeploymentTemplate(context contextpkg.Context, tx *sql.Tx, deployment *backend.Deployment) error {
	if deployment.TemplateID != "" {
		if template, err := self.getTemplateTx(context, tx, deployment.TemplateID); err == nil {
			deployment.MergeTemplate(template)
		} else {
			return err
		}
	}

	return nil
}

func (self *SQLBackend) updateDeploymentMetadata(context contextpkg.Context, tx *sql.Tx, deployment *backend.Deployment) error {
	deleteDeploymentMetadata := tx.StmtContext(context, self.statements.PreparedDeleteDeploymentMetadata)
	if _, err := deleteDeploymentMetadata.ExecContext(context, deployment.DeploymentID); err != nil {
		return err
	}

	return self.upsertDeploymentMetadata(context, tx, deployment)
}

func (self *SQLBackend) upsertDeploymentMetadata(context contextpkg.Context, tx *sql.Tx, deployment *backend.Deployment) error {
	upsertDeploymentMetadata := tx.StmtContext(context, self.statements.PreparedUpsertDeploymentMetadata)
	for key, value := range deployment.Metadata {
		if _, err := upsertDeploymentMetadata.ExecContext(context, deployment.DeploymentID, key, value); err != nil {
			return err
		}
	}

	return nil
}

func (self *SQLBackend) upsertTemplateDeployment(context contextpkg.Context, tx *sql.Tx, deployment *backend.Deployment) error {
	if deployment.TemplateID != "" {
		upsertTemplateDeployment := tx.StmtContext(context, self.statements.PreparedUpsertTemplateDeployment)
		if _, err := upsertTemplateDeployment.ExecContext(context, deployment.TemplateID, deployment.DeploymentID); err != nil {
			self.rollback(tx)
			return err
		}
	}

	return nil
}

func (self *SQLBackend) upsertSiteDeployment(context contextpkg.Context, tx *sql.Tx, deployment *backend.Deployment) error {
	if deployment.SiteID != "" {
		upsertSiteDeployment := tx.StmtContext(context, self.statements.PreparedUpsertSiteDeployment)
		if _, err := upsertSiteDeployment.ExecContext(context, deployment.SiteID, deployment.DeploymentID); err != nil {
			self.rollback(tx)
			return err
		}
	}

	return nil
}

func (self *SQLBackend) hasModificationExpired(modificationTimestamp *int64) bool {
	if modificationTimestamp == nil {
		return true
	}
	delta := time.Now().UnixMicro() - *modificationTimestamp
	return delta > self.maxModificationDuration
}
