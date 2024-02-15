package sql

import (
	contextpkg "context"
	"database/sql"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetSite(context contextpkg.Context, site *backend.Site) error {
	if tx, err := self.db.BeginTx(context, nil); err == nil {
		if err := self.mergeSiteTemplate(context, tx, site); err != nil {
			self.rollback(tx)
			return err
		}

		if resources, err := self.encodeResources(site.Resources); err == nil {
			upsertSite := tx.StmtContext(context, self.statements.PreparedUpsertSite)
			if _, err := upsertSite.ExecContext(context, site.SiteID, nilIfEmptyString(site.TemplateID), resources); err == nil {
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
	defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

	if rows.Next() {
		var templateId *string
		var resources, metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&resources, &templateId, &metadataJson, &deploymentIdsJson); err == nil {
			return self.newSite(siteId, templateId, metadataJson, deploymentIdsJson, resources)
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
func (self *SQLBackend) ListSites(context contextpkg.Context, listSites backend.ListSites) (util.Results[backend.SiteInfo], error) {
	sql := self.statements.SelectSites
	var args SqlArgs
	var with SqlWith
	var where SqlWhere

	for _, pattern := range listSites.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("sites.site_id ~ " + pattern)
	}

	for _, pattern := range listSites.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("template_id ~ " + pattern)
	}

	if listSites.MetadataPatterns != nil {
		for key, pattern := range listSites.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("SELECT site_id FROM sites_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")",
				"sites", "site_id")
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
			var metadataJson, deploymentIdsJson []byte
			if err := rows.Scan(&siteId, &templateId, &metadataJson, &deploymentIdsJson); err == nil {
				if siteInfo, err := self.newSiteInfo(siteId, templateId, metadataJson, deploymentIdsJson); err == nil {
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

// Utils

func (self *SQLBackend) newSiteInfo(siteId string, templateId *string, metadataJson []byte, deploymentIdsJson []byte) (backend.SiteInfo, error) {
	siteInfo := backend.SiteInfo{
		SiteID:   siteId,
		Metadata: make(map[string]string),
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

func (self *SQLBackend) newSite(siteId string, templateId *string, metadataJson []byte, deploymentIdsJson []byte, resources []byte) (*backend.Site, error) {
	if siteInfo, err := self.newSiteInfo(siteId, templateId, metadataJson, deploymentIdsJson); err == nil {
		site := backend.Site{SiteInfo: siteInfo}
		if site.Resources, err = self.decodeResources(resources); err == nil {
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
