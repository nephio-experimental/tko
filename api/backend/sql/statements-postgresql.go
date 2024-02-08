package sql

import (
	"database/sql"

	"github.com/tliron/commonlog"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresqlStatements(db *sql.DB, log commonlog.Logger) *Statements {
	return &Statements{
		db:  db,
		log: log,

		InsertTemplate: `
			INSERT INTO templates (template_id, resources)
			VALUES ($1, $2)
			ON CONFLICT (template_id)
				DO UPDATE SET
				resources = $2
		`,
		InsertTemplateMetadata: `
			INSERT INTO templates_metadata (template_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (template_id, key)
				DO UPDATE SET
				value = $3
		`,
		InsertTemplateDeployment: `
			INSERT INTO templates_deployments (template_id, deployment_id)
			VALUES ($1, $2)
			ON CONFLICT (template_id, deployment_id)
				DO NOTHING
		`,
		SelectTemplate: `
			SELECT resources, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM templates
			LEFT JOIN templates_metadata ON templates.template_id = templates_metadata.template_id
			LEFT JOIN templates_deployments ON templates.template_id = templates_deployments.template_id
			WHERE templates.template_id = $1
			GROUP BY templates.template_id
		`,
		SelectTemplateResources: `SELECT resources FROM templates WHERE template_id = $1`,
		DeleteTemplate:          `DELETE FROM templates WHERE template_id = $1`,
		SelectTemplates: `
			SELECT templates.template_id, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM templates
			LEFT JOIN templates_metadata ON templates.template_id = templates_metadata.template_id
			LEFT JOIN templates_deployments ON templates.template_id = templates_deployments.template_id
			GROUP BY templates.template_id
		`,

		InsertSite: `
			INSERT INTO sites (site_id, template_id, resources)
			VALUES ($1, $2, $3)
			ON CONFLICT (site_id)
				DO UPDATE SET
				template_id = $2, resources = $3
		`,
		InsertSiteMetadata: `
			INSERT INTO sites_metadata (site_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (site_id, key)
				DO UPDATE SET
				value = $3
		`,
		InsertSiteDeployment: `
			INSERT INTO sites_deployments (site_id, deployment_id)
			VALUES ($1, $2)
			ON CONFLICT (site_id, deployment_id)
				DO NOTHING
		`,
		SelectSite: `
			SELECT resources, template_id, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM sites
			LEFT JOIN sites_metadata ON sites.site_id = sites_metadata.site_id
			LEFT JOIN sites_deployments ON sites.site_id = sites_deployments.site_id
			WHERE sites.site_id = $1
			GROUP BY sites.site_id
		`,
		DeleteSite: `DELETE FROM sites WHERE site_id = $1`,
		SelectSites: `
			SELECT sites.site_id, template_id, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM sites
			LEFT JOIN sites_metadata ON sites.site_id = sites_metadata.site_id
			LEFT JOIN sites_deployments ON sites.site_id = sites_deployments.site_id
			GROUP BY sites.site_id
		`,

		InsertDeployment: `
			INSERT INTO deployments (deployment_id, parent_deployment_id, template_id, site_id, prepared, approved, resources)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (deployment_id)
				DO UPDATE SET
				parent_deployment_id = $2, template_id = $3, site_id = $4, prepared = $5, approved = $6, resources = $7
		`,
		UpdateDeployment: `
			UPDATE deployments
			SET template_id = $1, site_id = $2, prepared = $3, approved = $4, resources = $5, modification_token = NULL, modification_timestamp = 0
			WHERE deployment_id = $6
		`,
		SelectDeployment:                 `SELECT parent_deployment_id, template_id, site_id, prepared, approved, resources FROM deployments WHERE deployment_id = $1`,
		SelectDeploymentWithModificaiton: `SELECT parent_deployment_id, template_id, site_id, prepared, approved, resources, modification_token, modification_timestamp FROM deployments WHERE deployment_id = $1`,
		SelectDeploymentByModification:   `SELECT deployment_id, parent_deployment_id, template_id, site_id, prepared, approved, modification_timestamp FROM deployments WHERE modification_token = $1`,
		UpdateDeploymentModification: `
			UPDATE deployments
			SET modification_token = $1, modification_timestamp = $2
			WHERE deployment_id = $3
		`,
		ResetDeploymentModification: `
			UPDATE deployments
			SET modification_token = NULL, modification_timestamp = 0
			WHERE modification_token = $1
		`,
		DeleteDeployment: `DELETE FROM deployments WHERE deployment_id = $1`,
		SelectDeployments: `
			SELECT deployment_id, parent_deployment_id, deployments.template_id, deployments.site_id, prepared, approved
			FROM deployments
		`,

		InsertPlugin: `
			INSERT INTO plugins (type, "group", version, kind, executor, arguments, properties)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (type, "group", version, kind)
				DO UPDATE SET
				executor = $5, arguments = $6, properties = $7
		`,
		SelectPlugin:  `SELECT executor, arguments, properties FROM plugins WHERE type = $1 AND "group" = $2 AND version = $3 AND kind = $4`,
		DeletePlugin:  `DELETE FROM plugins WHERE type = $1 AND "group" = $2 AND version = $3 AND kind = $4`,
		SelectPlugins: `SELECT type, "group", version, kind, executor, arguments, properties FROM plugins`,

		CreateTemplates: `
			CREATE TABLE IF NOT EXISTS templates (
				template_id TEXT NOT NULL PRIMARY KEY,
				resources BYTEA
			)
		`,
		CreateTemplatesMetadata: `
			CREATE TABLE IF NOT EXISTS templates_metadata (
				template_id TEXT NOT NULL,
				key TEXT NOT NULL,
				value TEXT NOT NULL,
				UNIQUE (template_id, key),
				CONSTRAINT fk_template_id
					FOREIGN KEY (template_id)
					REFERENCES templates (template_id) ON DELETE CASCADE
			)
		`,
		CreateTemplatesMetadataIndex: `CREATE INDEX IF NOT EXISTS templates_metadata_key ON templates_metadata (key)`,
		CreateTemplatesDeployments: `
			CREATE TABLE IF NOT EXISTS templates_deployments (
				template_id TEXT NOT NULL,
				deployment_id TEXT NOT NULL,
				UNIQUE (template_id, deployment_id),
				CONSTRAINT fk_template_id
					FOREIGN KEY (template_id)
					REFERENCES templates (template_id) ON DELETE CASCADE,
				CONSTRAINT fk_deployment_id
					FOREIGN KEY (deployment_id)
					REFERENCES deployments (deployment_id) ON DELETE CASCADE
			)
		`,

		CreateSites: `
			CREATE TABLE IF NOT EXISTS sites (
				site_id TEXT NOT NULL PRIMARY KEY,
				resources BYTEA,
				template_id TEXT,
				CONSTRAINT fk_template_id
					FOREIGN KEY (template_id)
					REFERENCES templates (template_id) ON DELETE SET NULL
			)
		`,
		CreateSitesMetadata: `
			CREATE TABLE IF NOT EXISTS sites_metadata (
				site_id TEXT NOT NULL,
				key TEXT NOT NULL,
				value TEXT NOT NULL,
				UNIQUE (site_id, key),
				CONSTRAINT fk_site_id
					FOREIGN KEY (site_id)
					REFERENCES sites (site_id) ON DELETE CASCADE
			)
		`,
		CreateSitesMetadataIndex: `CREATE INDEX IF NOT EXISTS sites_metadata_key ON sites_metadata (key)`,
		CreateSitesDeployments: `
			CREATE TABLE IF NOT EXISTS sites_deployments (
				site_id TEXT NOT NULL,
				deployment_id TEXT NOT NULL,
				UNIQUE (site_id, deployment_id),
				CONSTRAINT fk_site_id
					FOREIGN KEY (site_id)
					REFERENCES sites (site_id) ON DELETE CASCADE,
				CONSTRAINT fk_deployment_id
					FOREIGN KEY (deployment_id)
					REFERENCES deployments (deployment_id) ON DELETE CASCADE
			)
		`,

		CreateDeployments: `
			CREATE TABLE IF NOT EXISTS deployments (
				deployment_id TEXT NOT NULL PRIMARY KEY,
				resources BYTEA,
				parent_deployment_id TEXT,
				template_id TEXT,
				site_id TEXT,
				prepared BOOLEAN,
				approved BOOLEAN,
				modification_token TEXT,
				modification_timestamp BIGINT,
				CONSTRAINT fk_parent_deployment_id
					FOREIGN KEY (parent_deployment_id)
					REFERENCES deployments (deployment_id) ON DELETE CASCADE,
				CONSTRAINT fk_template_id
					FOREIGN KEY (template_id)
					REFERENCES templates (template_id) ON DELETE SET NULL,
				CONSTRAINT fk_site_id
					FOREIGN KEY (site_id)
					REFERENCES sites (site_id) ON DELETE SET NULL
			)
		`,
		CreateDeploymentsPreparedIndex:     `CREATE INDEX IF NOT EXISTS deployments_prepared ON deployments (prepared)`,
		CreateDeploymentsApprovedIndex:     `CREATE INDEX IF NOT EXISTS deployments_approved ON deployments (approved)`,
		CreateDeploymentsModificationIndex: `CREATE INDEX IF NOT EXISTS deployments_modification ON deployments (modification_token)`,

		CreatePlugins: `
			CREATE TABLE IF NOT EXISTS plugins (
				type TEXT NOT NULL,
				"group" TEXT NOT NULL,
				version TEXT NOT NULL,
				kind TEXT NOT NULL,
				executor TEXT NOT NULL,
				arguments TEXT,
				properties TEXT,
				PRIMARY KEY (type, "group", version, kind)
			)
		`,

		DropTemplates:                    `DROP TABLE IF EXISTS templates`,
		DropTemplatesMetadata:            `DROP TABLE IF EXISTS templates_metadata`,
		DropTemplatesMetadataIndex:       `DROP INDEX IF EXISTS templates_metadata_key`,
		DropTemplatesDeployments:         `DROP TABLE IF EXISTS templates_deployments`,
		DropSites:                        `DROP TABLE IF EXISTS sites`,
		DropSitesMetadata:                `DROP TABLE IF EXISTS sites_metadata`,
		DropSitesMetadataIndex:           `DROP INDEX IF EXISTS sites_metadata_key`,
		DropSitesDeployments:             `DROP TABLE IF EXISTS sites_deployments`,
		DropDeployments:                  `DROP TABLE IF EXISTS deployments`,
		DropDeploymentsPreparedIndex:     `DROP INDEX IF EXISTS deployments_prepared`,
		DropDeploymentsApprovedIndex:     `DROP INDEX IF EXISTS deployments_approved`,
		DropDeploymentsModificationIndex: `DROP INDEX IF EXISTS deployments_modification`,
		DropPlugins:                      `DROP TABLE IF EXISTS plugins`,
	}
}
