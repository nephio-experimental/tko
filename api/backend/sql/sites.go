package sql

import (
	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
)

// ([backend.Backend] interface)
func (self *SqlBackend) SetSite(site *backend.Site) error {
	site = site.Clone()
	if site.Metadata == nil {
		site.Metadata = make(map[string]string)
	}

	site.Update()

	if tx, err := self.db.Begin(); err == nil {
		if site.TemplateID != "" {
			if resources, err := self.getTemplateResources(tx, site.TemplateID); err == nil {
				// Merge our resources over template resources
				resources = util.MergeResources(resources, site.Resources)

				site.Resources = resources
			} else {
				return err
			}
		}

		if resources, err := self.encodeResources(site.Resources); err == nil {
			if _, err := tx.Exec(self.sql.InsertSite, site.SiteID, nilIfEmptyString(site.TemplateID), resources); err == nil {
				if len(site.Metadata) > 0 {
					if insertSiteMetadata, err := tx.Prepare(self.sql.InsertSiteMetadata); err == nil {
						for key, value := range site.Metadata {
							if _, err := insertSiteMetadata.Exec(site.SiteID, key, value); err != nil {
								insertSiteMetadata.Close()
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return err
							}
						}
						insertSiteMetadata.Close()
					} else {
						if err := tx.Rollback(); err != nil {
							self.log.Error(err.Error())
						}
						return err
					}
				}

				if len(site.DeploymentIDs) > 0 {
					if insertSiteDeployment, err := tx.Prepare(self.sql.InsertSiteDeployment); err == nil {
						for _, deploymentId := range site.DeploymentIDs {
							if _, err := insertSiteDeployment.Exec(site.SiteID, deploymentId); err != nil {
								insertSiteDeployment.Close()
								if err := tx.Rollback(); err != nil {
									self.log.Error(err.Error())
								}
								return err
							}
						}
						insertSiteDeployment.Close()
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
func (self *SqlBackend) GetSite(siteId string) (*backend.Site, error) {
	rows, err := self.sql.PreparedSelectSite.Query(siteId)
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
func (self *SqlBackend) DeleteSite(siteId string) error {
	if result, err := self.sql.PreparedDeleteSite.Exec(siteId); err == nil {
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
func (self *SqlBackend) ListSites(siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]backend.SiteInfo, error) {
	sql := self.sql.SelectSites
	var args SqlArgs
	var where SqlWhere
	var with SqlWith

	for _, pattern := range siteIdPatterns {
		pattern = args.Add(backend.IdPatternRE(pattern))
		where.Add("(sites.site_id ~ " + pattern + ")")
	}

	for _, pattern := range templateIdPatterns {
		pattern = args.Add(backend.IdPatternRE(pattern))
		where.Add("(template_id ~ " + pattern + ")")
	}

	if metadataPatterns != nil {
		for key, pattern := range metadataPatterns {
			key = args.Add(key)
			pattern = args.Add(backend.PatternRE(pattern))
			with.Add("sites", "site_id", "SELECT site_id FROM sites_metadata WHERE (key = "+key+") AND (value ~ "+pattern+")")
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

	return siteInfos, nil
}
