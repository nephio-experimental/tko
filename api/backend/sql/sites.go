package sql

import (
	contextpkg "context"
	"database/sql"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/tliron/commonlog"
)

// ([backend.Backend] interface)
func (self *SQLBackend) SetSite(context contextpkg.Context, site *backend.Site) error {
	site = site.Clone()
	if site.Metadata == nil {
		site.Metadata = make(map[string]string)
	}

	site.Update()

	if tx, err := self.db.BeginTx(context, nil); err == nil {
		if err := self.mergeSite(context, tx, site); err != nil {
			if err := tx.Rollback(); err != nil {
				self.log.Error(err.Error())
			}
			return err
		}

		if resources, err := self.encodeResources(site.Resources); err == nil {
			insertSite := tx.StmtContext(context, self.statements.PreparedInsertSite)
			if _, err := insertSite.ExecContext(context, site.SiteID, nilIfEmptyString(site.TemplateID), resources); err == nil {
				insertSiteMetadata := tx.StmtContext(context, self.statements.PreparedInsertSiteMetadata)
				for key, value := range site.Metadata {
					if _, err := insertSiteMetadata.ExecContext(context, site.SiteID, key, value); err != nil {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return err
					}
				}

				insertSiteDeployment := tx.StmtContext(context, self.statements.PreparedInsertSiteDeployment)
				for _, deploymentId := range site.DeploymentIDs {
					if _, err := insertSiteDeployment.ExecContext(context, site.SiteID, deploymentId); err != nil {
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
func (self *SQLBackend) GetSite(context contextpkg.Context, siteId string) (*backend.Site, error) {
	rows, err := self.statements.PreparedSelectSite.QueryContext(context, siteId)
	if err != nil {
		return nil, err
	}
	defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

	if rows.Next() {
		var resources []byte
		var templateId *string
		var metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&resources, &templateId, &metadataJson, &deploymentIdsJson); err == nil {
			site := backend.Site{
				SiteInfo: backend.SiteInfo{
					SiteID:   siteId,
					Metadata: make(map[string]string),
				},
			}

			if site.Resources == nil {
				if site.Resources, err = self.decodeResources(resources); err != nil {
					return nil, err
				}
			}

			if templateId != nil {
				site.TemplateID = *templateId
			}

			if err := jsonUnmarshallMapEntries(metadataJson, site.Metadata); err != nil {
				return nil, err
			}

			if err := jsonUnmarshallArray(deploymentIdsJson, &site.DeploymentIDs); err != nil {
				return nil, err
			}

			return &site, nil
		} else {
			return nil, err
		}
	}
	return nil, backend.NewNotFoundErrorf("site: %s", siteId)
}

// ([backend.Backend] interface)
func (self *SQLBackend) DeleteSite(context contextpkg.Context, siteId string) error {
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
func (self *SQLBackend) ListSites(context contextpkg.Context, listSites backend.ListSites) (backend.SiteInfoStream, error) {
	sql := self.statements.SelectSites
	var args SqlArgs
	var where SqlWhere
	var with SqlWith

	for _, pattern := range listSites.SiteIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("(sites.site_id ~ " + pattern + ")")
	}

	for _, pattern := range listSites.TemplateIDPatterns {
		pattern = args.Add(backend.IDPatternRE(pattern))
		where.Add("(template_id ~ " + pattern + ")")
	}

	if listSites.MetadataPatterns != nil {
		for key, pattern := range listSites.MetadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("sites", "site_id", "SELECT site_id FROM sites_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
		}
	}

	sql = where.Apply(sql)
	sql = with.Apply(sql)
	self.log.Debugf("generated SQL: %s", sql)

	rows, err := self.db.QueryContext(context, sql, args.Args...)
	if err != nil {
		return nil, err
	}
	defer commonlog.CallAndLogError(rows.Close, "rows.Close", self.log)

	var siteInfos []backend.SiteInfo
	for rows.Next() {
		var siteId string
		var templateId *string
		var metadataJson, deploymentIdsJson []byte
		if err := rows.Scan(&siteId, &templateId, &metadataJson, &deploymentIdsJson); err == nil {
			siteInfo := backend.SiteInfo{
				SiteID:   siteId,
				Metadata: make(map[string]string),
			}

			if templateId != nil {
				siteInfo.TemplateID = *templateId
			}

			if err := jsonUnmarshallMapEntries(metadataJson, siteInfo.Metadata); err != nil {
				return nil, err
			}

			if err := jsonUnmarshallArray(deploymentIdsJson, &siteInfo.DeploymentIDs); err != nil {
				return nil, err
			}

			siteInfos = append(siteInfos, siteInfo)
		} else {
			return nil, err
		}
	}

	return backend.NewSiteInfoSliceStream(siteInfos), nil
}

func (self *SQLBackend) mergeSite(context contextpkg.Context, tx *sql.Tx, site *backend.Site) error {
	if site.TemplateID != "" {
		if template, err := self.getTemplateTx(context, tx, site.TemplateID); err == nil {
			site.MergeTemplate(template)
		} else {
			return err
		}
	}

	return nil
}
