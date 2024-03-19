package sql

import (
	contextpkg "context"
	"database/sql"
	"time"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetSite(context contextpkg.Context, site *backend.Site) error {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		if err := self.mergeSiteTemplate(context, tx, site); err != nil {
			self.rollback(tx)
			return err
		}

		var package_ []byte
		var err error
		if package_, err = self.encodePackage(site.Package); err != nil {
			self.rollback(tx)
			return err
		}

		site.Updated = time.Now().UTC()
		upsertSite := tx.StmtContext(context, self.statements.PreparedUpsertSite)
		if _, err := upsertSite.ExecContext(context, site.SiteID, nilIfEmptyString(site.TemplateID), site.Updated, package_); err == nil {
			if err := self.updateSiteMetadata(context, tx, site); err != nil {
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
func (self *SQLBackend) GetSite(context contextpkg.Context, siteId string) (*backend.Site, error) {
	rows, err := self.statements.PreparedSelectSite.QueryContext(context, siteId)
	if err != nil {
		return nil, err
	}
	defer self.closeRows(rows)

	if rows.Next() {
		var templateId *string
		var updated time.Time
		var package_, metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&templateId, &updated, &package_, &metadataJson, &deploymentIdsJson); err == nil {
			return self.newSite(siteId, templateId, updated, metadataJson, deploymentIdsJson, package_)
		} else {
			return nil, err
		}
	}

	return nil, backend.NewNotFoundErrorf("site: %s", siteId)
}

// ([backend.Backend] interface)
func (self *SQLBackend) DeleteSite(context contextpkg.Context, siteId string) error {
	// Will cascade delete sites_metadata, templates_deployments, sites_deployments
	if result, err := self.statements.PreparedDeleteSite.ExecContext(context, siteId); err == nil {
		if count, err := result.RowsAffected(); err == nil {
			if count == 0 {
				return backend.NewNotFoundErrorf("site: %s", siteId)
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
func (self *SQLBackend) ListSites(context contextpkg.Context, selectSites backend.SelectSites, window backend.Window) (util.Results[backend.SiteInfo], error) {
	sql := self.statements.SelectSites
	var args SqlArgs
	var with SqlWith
	var where SqlWhere

	args.AddValue(window.Offset)
	args.AddValue(window.Limit())

	for _, pattern := range selectSites.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`sites.site_id ~ ` + pattern)
	}

	for _, pattern := range selectSites.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`template_id ~ ` + pattern)
	}

	if len(selectSites.MetadataPatterns) > 0 {
		for key, pattern := range selectSites.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add(`SELECT site_id FROM sites_metadata WHERE (key = `+key+`) AND (value ~ `+pattern+`)`,
				`sites`, `site_id`)
		}
	}

	sql = with.Apply(sql)
	sql = where.Apply(sql)
	self.log.Debugf("generated SQL:\n%s", sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}

	stream := util.NewResultsStream[backend.SiteInfo](func() {
		self.closeRows(rows)
	})

	go func() {
		for rows.Next() {
			var siteId string
			var templateId *string
			var updated time.Time
			var metadataJson, deploymentIdsJson []byte
			if err := rows.Scan(&siteId, &templateId, &updated, &metadataJson, &deploymentIdsJson); err == nil {
				if siteInfo, err := self.newSiteInfo(siteId, templateId, updated, metadataJson, deploymentIdsJson); err == nil {
					stream.Send(siteInfo)
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
func (self *SQLBackend) PurgeSites(context contextpkg.Context, selectSites backend.SelectSites) error {
	sql := self.statements.DeleteSites
	var args SqlArgs
	var where SqlWhere

	for _, pattern := range selectSites.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`sites.site_id ~ ` + pattern)
	}

	for _, pattern := range selectSites.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add(`template_id ~ ` + pattern)
	}

	if len(selectSites.MetadataPatterns) > 0 {
		where.Add(`sites.site_id = sites_metadata.site_id`)
		for key, pattern := range selectSites.MetadataPatterns {
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

func (self *SQLBackend) newSiteInfo(siteId string, templateId *string, updated time.Time, metadataJson []byte, deploymentIdsJson []byte) (backend.SiteInfo, error) {
	siteInfo := backend.SiteInfo{
		SiteID:   siteId,
		Metadata: make(map[string]string),
		Updated:  updated,
	}

	if templateId != nil {
		siteInfo.TemplateID = *templateId
	}

	if err := jsonUnmarshallStringMapEntries(metadataJson, siteInfo.Metadata); err != nil {
		return backend.SiteInfo{}, err
	}

	if err := jsonUnmarshallStringArray(deploymentIdsJson, &siteInfo.DeploymentIDs); err != nil {
		return backend.SiteInfo{}, err
	}

	return siteInfo, nil
}

func (self *SQLBackend) newSite(siteId string, templateId *string, updated time.Time, metadataJson []byte, deploymentIdsJson []byte, package_ []byte) (*backend.Site, error) {
	if siteInfo, err := self.newSiteInfo(siteId, templateId, updated, metadataJson, deploymentIdsJson); err == nil {
		site := backend.Site{SiteInfo: siteInfo}
		if site.Package, err = self.decodePackage(package_); err == nil {
			return &site, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *SQLBackend) mergeSiteTemplate(context contextpkg.Context, tx *sql.Tx, site *backend.Site) error {
	if site.TemplateID != "" {
		if template, err := self.getTemplateTx(context, tx, site.TemplateID); err == nil {
			site.MergeTemplate(template)
		} else {
			return err
		}
	}

	return nil
}

func (self *SQLBackend) updateSiteMetadata(context contextpkg.Context, tx *sql.Tx, site *backend.Site) error {
	deleteSiteMetadata := tx.StmtContext(context, self.statements.PreparedDeleteSiteMetadata)
	if _, err := deleteSiteMetadata.ExecContext(context, site.SiteID); err != nil {
		return err
	}

	upsertSiteMetadata := tx.StmtContext(context, self.statements.PreparedUpsertSiteMetadata)
	for key, value := range site.Metadata {
		if _, err := upsertSiteMetadata.ExecContext(context, site.SiteID, key, value); err != nil {
			return err
		}
	}

	return nil
}
