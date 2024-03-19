package sql

import (
	contextpkg "context"
	"database/sql"
	"time"

	"github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
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
		deployment.UpdateFromPackage(true)

		var package_ []byte
		var err error
		if package_, err = self.encodePackage(deployment.Package); err != nil {
			self.rollback(tx)
			return err
		}

		deployment.DeploymentID = backend.NewID()
		now := time.Now().UTC()
		deployment.Created = now
		deployment.Updated = now

		insertDeployment := tx.StmtContext(context, self.statements.PreparedInsertDeployment)
		if _, err := insertDeployment.ExecContext(context, deployment.DeploymentID, nilIfEmptyString(deployment.ParentDeploymentID), nilIfEmptyString(deployment.TemplateID), nilIfEmptyString(deployment.SiteID), deployment.Created, deployment.Updated, deployment.Prepared, deployment.Approved, package_); err != nil {
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
	defer self.closeRows(rows)

	if rows.Next() {
		var parentDeploymentId, templateId, siteId *string
		var created, updated time.Time
		var prepared, approved bool
		var metadataJson, package_ []byte
		if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &metadataJson, &created, &updated, &prepared, &approved, &package_); err == nil {
			return self.newDeployment(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, created, updated, prepared, approved, package_)
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
func (self *SQLBackend) ListDeployments(context contextpkg.Context, selectDeployments backend.SelectDeployments, window backend.Window) (util.Results[backend.DeploymentInfo], error) {
	sql := self.statements.SelectDeployments
	var with SqlWith
	var where SqlWhere
	var args SqlArgs

	args.AddValue(window.Offset)
	args.AddValue(window.Limit())

	if (selectDeployments.ParentDeploymentID != nil) && (*selectDeployments.ParentDeploymentID != "") {
		where.Add(`parent_deployment_id = ` + args.Add(selectDeployments.ParentDeploymentID))
	}

	if len(selectDeployments.MetadataPatterns) > 0 {
		for key, pattern := range selectDeployments.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add(`SELECT deployment_id FROM deployments_metadata WHERE (key = `+key+`) AND (value ~ `+pattern+`)`,
				`deployments`, `deployment_id`)
		}
	}

	for _, pattern := range selectDeployments.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`deployments.template_id ~ ` + pattern)
	}

	if len(selectDeployments.TemplateMetadataPatterns) > 0 {
		for key, pattern := range selectDeployments.TemplateMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add(`SELECT template_id FROM templates_metadata WHERE (key = `+key+`) AND (value ~ `+pattern+`)`,
				`deployments`, `template_id`)
		}
	}

	for _, pattern := range selectDeployments.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`deployments.site_id ~ ` + pattern)
	}

	if len(selectDeployments.SiteMetadataPatterns) > 0 {
		for key, pattern := range selectDeployments.SiteMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add(`SELECT site_id FROM sites_metadata WHERE (key = `+key+`) AND (value ~ `+pattern+`)`,
				`deployments`, `site_id`)
		}
	}

	if selectDeployments.Prepared != nil {
		switch *selectDeployments.Prepared {
		case true:
			where.Add(`prepared`)
		case false:
			where.Add(`NOT prepared`)
		}
	}

	if selectDeployments.Approved != nil {
		switch *selectDeployments.Approved {
		case true:
			where.Add(`approved`)
		case false:
			where.Add(`NOT approved`)
		}
	}

	sql = with.Apply(sql)
	sql = where.Apply(sql)
	self.log.Debugf("generated SQL:\n%s", sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}

	stream := util.NewResultsStream[backend.DeploymentInfo](func() {
		self.closeRows(rows)
	})

	go func() {
		for rows.Next() {
			var deploymentId string
			var parentDeploymentId, templateId, siteId *string
			var metadataJson []byte
			var created, updated time.Time
			var prepared, approved bool
			if err := rows.Scan(&deploymentId, &parentDeploymentId, &templateId, &siteId, &metadataJson, &created, &updated, &prepared, &approved); err == nil {
				if deploymentInfo, err := self.newDeploymentInfo(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, created, updated, prepared, approved); err == nil {
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
func (self *SQLBackend) PurgeDeployments(context contextpkg.Context, selectDeployments backend.SelectDeployments) error {
	sql := self.statements.DeleteDeployments
	var args SqlArgs
	var where SqlWhere

	if (selectDeployments.ParentDeploymentID != nil) && (*selectDeployments.ParentDeploymentID != "") {
		where.Add(`parent_deployment_id = ` + args.Add(selectDeployments.ParentDeploymentID))
	}

	if len(selectDeployments.MetadataPatterns) > 0 {
		where.Add(`deployments.deployment_id = deployments_metadata.deployment_id`)
		for key, pattern := range selectDeployments.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			where.Add(`deployments_metadata.key = ` + key)
			where.Add(`deployments_metadata.value ~ ` + pattern)
		}
	}

	for _, pattern := range selectDeployments.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`template_id ~ ` + pattern)
	}

	if len(selectDeployments.TemplateMetadataPatterns) > 0 {
		where.Add(`deployments.template_id = templates_metadata.template_id`)
		for key, pattern := range selectDeployments.TemplateMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			where.Add(`templates_metadata.key = ` + key)
			where.Add(`templates_metadata.value ~ ` + pattern)
		}
	}

	for _, pattern := range selectDeployments.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`deployments.site_id ~ ` + pattern)
	}

	if len(selectDeployments.SiteMetadataPatterns) > 0 {
		where.Add(`deployments.site_id = sites_metadata.site_id`)
		for key, pattern := range selectDeployments.TemplateMetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			where.Add(`sites_metadata.key = ` + key)
			where.Add(`sites_metadata.value ~ ` + pattern)
		}
	}

	if selectDeployments.Prepared != nil {
		switch *selectDeployments.Prepared {
		case true:
			where.Add(`prepared`)
		case false:
			where.Add(`NOT prepared`)
		}
	}

	if selectDeployments.Approved != nil {
		switch *selectDeployments.Approved {
		case true:
			where.Add(`approved`)
		case false:
			where.Add(`NOT approved`)
		}
	}

	sql = where.Apply(sql)
	self.log.Debugf("generated SQL:\n%s", sql)

	_, err := self.db.ExecContext(context, sql, args.Args...)
	return err
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
			var created, updated time.Time
			var prepared, approved bool
			var metadataJson, package_ []byte
			var modificationTimestamp *int64
			if err := rows.Scan(&parentDeploymentId, &templateId, &siteId, &metadataJson, &created, &updated, &prepared, &approved, &package_, &modificationToken, &modificationTimestamp); err == nil {
				self.closeRows(rows)

				available := (modificationToken == nil) || (*modificationToken == "")
				if !available {
					available = self.hasModificationExpired(modificationTimestamp)
				}

				if !available {
					self.rollback(tx)
					return "", nil, backend.NewBusyErrorf("deployment: %s", deploymentId)
				}

				if deployment, err := self.newDeployment(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, created, updated, prepared, approved, package_); err == nil {
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
func (self *SQLBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, package_ tkoutil.Package, validation *validationpkg.Validation) (string, error) {
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
			var created, updated time.Time
			var prepared, approved bool
			var modificationTimestamp *int64
			if err := rows.Scan(&deploymentId, &templateId, &siteId, &metadataJson, &created, &updated, &prepared, &approved, &modificationTimestamp); err == nil {
				self.closeRows(rows)

				if self.hasModificationExpired(modificationTimestamp) {
					self.rollback(tx)
					return "", backend.NewNotFoundErrorf("modification token: %s", modificationToken)
				}

				var deploymentInfo backend.DeploymentInfo
				if deploymentInfo, err = self.newDeploymentInfo(deploymentId, nil, templateId, siteId, metadataJson, created, updated, prepared, approved); err != nil {
					self.rollback(tx)
					return "", err
				}

				deployment := backend.Deployment{
					DeploymentInfo: deploymentInfo,
					Package:        package_,
				}

				originalTemplateId := deploymentInfo.TemplateID
				originalSiteId := deploymentInfo.SiteID
				originalMetadata := tkoutil.CloneStringMap(deployment.Metadata)
				deployment.UpdateFromPackage(false)

				if validation != nil {
					// Complete validation when fully prepared
					if err := validation.ValidatePackage(package_, deployment.Prepared); err != nil {
						self.rollback(tx)
						return "", err
					}
				}

				var package_ []byte
				var err error
				if package_, err = self.encodePackage(deployment.Package); err != nil {
					self.rollback(tx)
					return "", err
				}

				deployment.Updated = time.Now().UTC()

				updateDeployment := tx.StmtContext(context, self.statements.PreparedUpdateDeployment)
				if _, err := updateDeployment.ExecContext(context, deployment.DeploymentID, deployment.Updated, deployment.Prepared, deployment.Approved, package_); err != nil {
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

func (self *SQLBackend) newDeploymentInfo(deploymentId string, parentDeploymentId *string, templateId *string, siteId *string, metadataJson []byte, created time.Time, updated time.Time, prepared bool, approved bool) (backend.DeploymentInfo, error) {
	deploymentInfo := backend.DeploymentInfo{
		DeploymentID: deploymentId,
		Metadata:     make(map[string]string),
		Created:      created,
		Updated:      updated,
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

func (self *SQLBackend) newDeployment(deploymentId string, parentDeploymentId *string, templateId *string, siteId *string, metadataJson []byte, created time.Time, updated time.Time, prepared bool, approved bool, package_ []byte) (*backend.Deployment, error) {
	if deploymentInfo, err := self.newDeploymentInfo(deploymentId, parentDeploymentId, templateId, siteId, metadataJson, created, updated, prepared, approved); err == nil {
		deployment := backend.Deployment{DeploymentInfo: deploymentInfo}
		if deployment.Package, err = self.decodePackage(package_); err == nil {
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
