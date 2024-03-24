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

		// Templates

		CreateTemplates: CleanSQL(`
			CREATE TABLE IF NOT EXISTS templates (
				template_id TEXT NOT NULL PRIMARY KEY,
				updated TIMESTAMP,
				package BYTEA
			)
		`),
		DropTemplates: `DROP TABLE IF EXISTS templates`,
		CreateTemplatesMetadata: CleanSQL(`
			CREATE TABLE IF NOT EXISTS templates_metadata (
				template_id TEXT NOT NULL,
				key TEXT NOT NULL,
				value TEXT NOT NULL,
				UNIQUE (template_id, key),
				CONSTRAINT fk_template_id
					FOREIGN KEY (template_id)
					REFERENCES templates (template_id) ON DELETE CASCADE
			)
		`),
		DropTemplatesMetadata:        `DROP TABLE IF EXISTS templates_metadata`,
		CreateTemplatesMetadataIndex: `CREATE INDEX IF NOT EXISTS templates_metadata_index ON templates_metadata (key, value)`,
		DropTemplatesMetadataIndex:   `DROP INDEX IF EXISTS templates_metadata_index`,
		CreateTemplatesDeployments: CleanSQL(`
			CREATE TABLE IF NOT EXISTS templates_deployments (
				template_id TEXT NOT NULL,
				deployment_id TEXT NOT NULL,
				UNIQUE (deployment_id),
				CONSTRAINT fk_template_id
					FOREIGN KEY (template_id)
					REFERENCES templates (template_id) ON DELETE CASCADE,
				CONSTRAINT fk_deployment_id
					FOREIGN KEY (deployment_id)
					REFERENCES deployments (deployment_id) ON DELETE CASCADE
			)
		`),
		DropTemplatesDeployments: `DROP TABLE IF EXISTS templates_deployments`,

		UpsertTemplate: CleanSQL(`
			INSERT INTO templates (template_id, updated, package)
			VALUES ($1, $2, $3)
			ON CONFLICT (template_id)
				DO UPDATE SET
				updated = $2, package = $3
		`),
		UpsertTemplateMetadata: CleanSQL(`
			INSERT INTO templates_metadata (template_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (template_id, key)
				DO UPDATE SET
				value = $3
		`),
		UpsertTemplateDeployment: CleanSQL(`
			INSERT INTO templates_deployments (template_id, deployment_id)
			VALUES ($1, $2)
			ON CONFLICT (deployment_id)
				DO UPDATE SET
				template_id = $1
		`),
		SelectTemplate: CleanSQL(`
			SELECT updated, package, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM templates
			LEFT JOIN templates_metadata ON templates.template_id = templates_metadata.template_id
			LEFT JOIN templates_deployments ON templates.template_id = templates_deployments.template_id
			WHERE templates.template_id = $1
			GROUP BY templates.template_id
		`),
		DeleteTemplate:           `DELETE FROM templates WHERE template_id = $1`,
		DeleteTemplateMetadata:   `DELETE FROM templates_metadata WHERE template_id = $1`,
		DeleteTemplateDeployment: `DELETE FROM templates_deployments WHERE deployment_id = $1`,
		DeleteTemplates: CleanSQL(`
			DELETE FROM templates
			USING templates_metadata
		`),
		SelectTemplates: CleanSQL(`
			SELECT templates.template_id, updated, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM templates
			LEFT JOIN templates_metadata ON templates.template_id = templates_metadata.template_id
			LEFT JOIN templates_deployments ON templates.template_id = templates_deployments.template_id
			GROUP BY templates.template_id
			ORDER BY templates.template_id
			LIMIT $2 OFFSET $1
		`),

		// Sites

		CreateSites: CleanSQL(`
			CREATE TABLE IF NOT EXISTS sites (
				site_id TEXT NOT NULL PRIMARY KEY,
				template_id TEXT,
				updated TIMESTAMP,
				package BYTEA,
				CONSTRAINT fk_template_id
					FOREIGN KEY (template_id)
					REFERENCES templates (template_id) ON DELETE SET NULL
			)
		`),
		DropSites: `DROP TABLE IF EXISTS sites`,
		CreateSitesMetadata: CleanSQL(`
			CREATE TABLE IF NOT EXISTS sites_metadata (
				site_id TEXT NOT NULL,
				key TEXT NOT NULL,
				value TEXT NOT NULL,
				UNIQUE (site_id, key),
				CONSTRAINT fk_site_id
					FOREIGN KEY (site_id)
					REFERENCES sites (site_id) ON DELETE CASCADE
			)
		`),
		DropSitesMetadata:        `DROP TABLE IF EXISTS sites_metadata`,
		CreateSitesMetadataIndex: `CREATE INDEX IF NOT EXISTS sites_metadata_index ON sites_metadata (key, value)`,
		DropSitesMetadataIndex:   `DROP INDEX IF EXISTS sites_metadata_index`,
		CreateSitesDeployments: CleanSQL(`
			CREATE TABLE IF NOT EXISTS sites_deployments (
				site_id TEXT NOT NULL,
				deployment_id TEXT NOT NULL,
				UNIQUE (deployment_id),
				CONSTRAINT fk_site_id
					FOREIGN KEY (site_id)
					REFERENCES sites (site_id) ON DELETE CASCADE,
				CONSTRAINT fk_deployment_id
					FOREIGN KEY (deployment_id)
					REFERENCES deployments (deployment_id) ON DELETE CASCADE
			)
		`),
		DropSitesDeployments: `DROP TABLE IF EXISTS sites_deployments`,

		UpsertSite: CleanSQL(`
			INSERT INTO sites (site_id, template_id, updated, package)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (site_id)
				DO UPDATE SET
				template_id = $2, updated = $3, package = $4
		`),
		UpsertSiteMetadata: CleanSQL(`
			INSERT INTO sites_metadata (site_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (site_id, key)
				DO UPDATE SET
				value = $3
		`),
		UpsertSiteDeployment: CleanSQL(`
			INSERT INTO sites_deployments (site_id, deployment_id)
			VALUES ($1, $2)
			ON CONFLICT (deployment_id)
				DO UPDATE SET
				site_id = $1
		`),
		SelectSite: CleanSQL(`
			SELECT template_id, updated, package, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM sites
			LEFT JOIN sites_metadata ON sites.site_id = sites_metadata.site_id
			LEFT JOIN sites_deployments ON sites.site_id = sites_deployments.site_id
			WHERE sites.site_id = $1
			GROUP BY sites.site_id
		`),
		DeleteSite:           `DELETE FROM sites WHERE site_id = $1`,
		DeleteSiteMetadata:   `DELETE FROM sites_metadata WHERE site_id = $1`,
		DeleteSiteDeployment: `DELETE FROM sites_deployments WHERE deployment_id = $1`,
		DeleteSites: CleanSQL(`
			DELETE FROM sites
			USING sites_metadata
		`),
		SelectSites: CleanSQL(`
			SELECT sites.site_id, template_id, updated, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), JSON_AGG (DISTINCT deployment_id) FILTER (WHERE deployment_id IS NOT NULL)
			FROM sites
			LEFT JOIN sites_metadata ON sites.site_id = sites_metadata.site_id
			LEFT JOIN sites_deployments ON sites.site_id = sites_deployments.site_id
			GROUP BY sites.site_id
			ORDER BY sites.site_id
			LIMIT $2 OFFSET $1
		`),

		// Deployments

		CreateDeployments: CleanSQL(`
			CREATE TABLE IF NOT EXISTS deployments (
				deployment_id TEXT NOT NULL PRIMARY KEY,
				parent_deployment_id TEXT,
				template_id TEXT,
				site_id TEXT,
				created TIMESTAMP,
				updated TIMESTAMP,
				prepared BOOLEAN,
				approved BOOLEAN,
				package BYTEA,
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
		`),
		DropDeployments: `DROP TABLE IF EXISTS deployments`,
		CreateDeploymentsMetadata: CleanSQL(`
			CREATE TABLE IF NOT EXISTS deployments_metadata (
				deployment_id TEXT NOT NULL,
				key TEXT NOT NULL,
				value TEXT NOT NULL,
				UNIQUE (deployment_id, key),
				CONSTRAINT fk_deployment_id
					FOREIGN KEY (deployment_id)
					REFERENCES deployments (deployment_id) ON DELETE CASCADE
			)
		`),
		DropDeploymentsMetadata:            `DROP TABLE IF EXISTS deployments_metadata`,
		CreateDeploymentsMetadataIndex:     `CREATE INDEX IF NOT EXISTS deployments_metadata_index ON deployments_metadata (key, value)`,
		DropDeploymentsMetadataIndex:       `DROP INDEX IF EXISTS deployments_metadata_index`,
		CreateDeploymentsPreparedIndex:     `CREATE INDEX IF NOT EXISTS deployments_prepared_index ON deployments (prepared)`,
		DropDeploymentsPreparedIndex:       `DROP INDEX IF EXISTS deployments_prepared_index`,
		CreateDeploymentsApprovedIndex:     `CREATE INDEX IF NOT EXISTS deployments_approved_index ON deployments (approved)`,
		DropDeploymentsApprovedIndex:       `DROP INDEX IF EXISTS deployments_approved_index`,
		CreateDeploymentsModificationIndex: `CREATE INDEX IF NOT EXISTS deployments_modification_index ON deployments (modification_token)`,
		DropDeploymentsModificationIndex:   `DROP INDEX IF EXISTS deployments_modification_index`,

		InsertDeployment: CleanSQL(`
			INSERT INTO deployments (deployment_id, parent_deployment_id, template_id, site_id, created, updated, prepared, approved, package)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`),
		UpdateDeployment: CleanSQL(`
			UPDATE deployments
			SET updated = $2, prepared = $3, approved = $4, package = $5, modification_token = NULL, modification_timestamp = 0
			WHERE deployment_id = $1
		`),
		UpsertDeploymentMetadata: CleanSQL(`
			INSERT INTO deployments_metadata (deployment_id, key, value)
			VALUES ($1, $2, $3)
			ON CONFLICT (deployment_id, key)
				DO UPDATE SET
				value = $3
		`),
		SelectDeployment: CleanSQL(`
			SELECT parent_deployment_id, template_id, site_id, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), created, updated, prepared, approved, package
			FROM deployments
			LEFT JOIN deployments_metadata ON deployments.deployment_id = deployments_metadata.deployment_id
			WHERE deployments.deployment_id = $1
			GROUP BY deployments.deployment_id
		`),
		SelectDeploymentWithModification: CleanSQL(`
			SELECT parent_deployment_id, template_id, site_id, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), created, updated, prepared, approved, package, modification_token, modification_timestamp
			FROM deployments
			LEFT JOIN deployments_metadata ON deployments.deployment_id = deployments_metadata.deployment_id
			WHERE deployments.deployment_id = $1
			GROUP BY deployments.deployment_id
		`),
		SelectDeploymentByModification: CleanSQL(`
			SELECT deployments.deployment_id, template_id, site_id, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), created, updated, prepared, approved, modification_timestamp
			FROM deployments
			LEFT JOIN deployments_metadata ON deployments.deployment_id = deployments_metadata.deployment_id
			WHERE modification_token = $1
			GROUP BY deployments.deployment_id
		`),
		UpdateDeploymentModification: CleanSQL(`
			UPDATE deployments
			SET modification_token = $1, modification_timestamp = $2
			WHERE deployment_id = $3
		`),
		ResetDeploymentModification: CleanSQL(`
			UPDATE deployments
			SET modification_token = NULL, modification_timestamp = 0
			WHERE modification_token = $1
		`),
		DeleteDeployment:         `DELETE FROM deployments WHERE deployment_id = $1`,
		DeleteDeploymentMetadata: `DELETE FROM deployments_metadata WHERE deployment_id = $1`,
		DeleteDeployments: CleanSQL(`
			DELETE FROM deployments
			USING deployments_metadata, templates_metadata, sites_metadata
		`),
		SelectDeployments: CleanSQL(`
			SELECT deployments.deployment_id, parent_deployment_id, deployments.template_id, deployments.site_id, JSON_AGG (ARRAY [key, value]) FILTER (WHERE key IS NOT NULL), created, updated, prepared, approved
			FROM deployments
			LEFT JOIN deployments_metadata ON deployments.deployment_id = deployments_metadata.deployment_id
			GROUP BY deployments.deployment_id
			ORDER BY deployments.deployment_id
			LIMIT $2 OFFSET $1
		`),

		// Plugins

		CreatePlugins: CleanSQL(`
			CREATE TABLE IF NOT EXISTS plugins (
				type TEXT NOT NULL,
				name TEXT NOT NULL,
				executor TEXT NOT NULL,
				arguments TEXT,
				properties TEXT,
				PRIMARY KEY (type, name)
			)
		`),
		DropPlugins:                `DROP TABLE IF EXISTS plugins`,
		CreatePluginsTypeIndex:     `CREATE INDEX IF NOT EXISTS plugins_type_index ON plugins (type)`,
		DropPluginsTypeIndex:       `DROP INDEX IF EXISTS plugins_type_index`,
		CreatePluginsExecutorIndex: `CREATE INDEX IF NOT EXISTS plugins_executor_index ON plugins (executor)`,
		DropPluginsExecutorIndex:   `DROP INDEX IF EXISTS plugins_executor_index`,
		CreatePluginsTriggers: CleanSQL(`
			CREATE TABLE IF NOT EXISTS plugins_triggers (
				plugin_type TEXT NOT NULL,
				plugin_name TEXT NOT NULL,
				"group" TEXT NOT NULL,
				version TEXT NOT NULL,
				kind TEXT NOT NULL,
				UNIQUE (plugin_type, plugin_name, "group", version, kind),
				CONSTRAINT fk_plugin_id
					FOREIGN KEY (plugin_type, plugin_name)
					REFERENCES plugins (type, name) ON DELETE CASCADE
			)
		`),
		DropPluginsTriggers:        `DROP TABLE IF EXISTS plugins_triggers`,
		CreatePluginsTriggersIndex: `CREATE INDEX IF NOT EXISTS plugins_triggers_index ON plugins_triggers ("group", version, kind)`,
		DropPluginsTriggersIndex:   `DROP INDEX IF EXISTS plugins_triggers_index`,

		UpsertPlugin: CleanSQL(`
			INSERT INTO plugins (type, name, executor, arguments, properties)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (type, name)
				DO UPDATE SET
				executor = $3, arguments = $4, properties = $5
		`),
		InsertPluginTrigger: CleanSQL(`
			INSERT INTO plugins_triggers (plugin_type, plugin_name, "group", version, kind)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (plugin_type, plugin_name, "group", version, kind)
				DO NOTHING
		`),
		SelectPlugin: CleanSQL(`
			SELECT executor, arguments, properties, JSON_AGG (ARRAY ["group", version, kind]) FILTER (WHERE "group" IS NOT NULL)
			FROM plugins
			LEFT JOIN plugins_triggers ON plugins.type = plugins_triggers.plugin_type AND plugins.name = plugins_triggers.plugin_name
			WHERE type = $1 AND name = $2
			GROUP BY type, name
		`),
		DeletePlugin:         `DELETE FROM plugins WHERE type = $1 AND name = $2`,
		DeletePluginTriggers: `DELETE FROM plugins_triggers WHERE plugin_type = $1 AND plugin_name = $2`,
		DeletePlugins: CleanSQL(`
			DELETE FROM plugins
			USING plugins_triggers
		`),
		SelectPlugins: CleanSQL(`
			SELECT plugins.type, plugins.name, executor, arguments, properties, JSON_AGG (ARRAY ["group", version, kind]) FILTER (WHERE "group" IS NOT NULL)
			FROM plugins
			LEFT JOIN plugins_triggers ON plugins.type = plugins_triggers.plugin_type AND plugins.name = plugins_triggers.plugin_name
			GROUP BY plugins.type, plugins.name
			ORDER BY plugins.type, plugins.name
			LIMIT $2 OFFSET $1
		`),
	}
}
