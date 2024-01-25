package sql

import (
	contextpkg "context"
	"database/sql"
	"fmt"

	"github.com/tliron/commonlog"
)

//
// Statements
//

type Statements struct {
	InsertTemplate           string
	InsertTemplateMetadata   string
	InsertTemplateDeployment string
	SelectTemplate           string
	SelectTemplateResources  string
	DeleteTemplate           string
	SelectTemplates          string

	InsertSite           string
	InsertSiteMetadata   string
	InsertSiteDeployment string
	SelectSite           string
	DeleteSite           string
	SelectSites          string

	InsertDeployment                 string
	UpdateDeployment                 string
	SelectDeployment                 string
	SelectDeploymentWithModificaiton string
	SelectDeploymentByModification   string
	UpdateDeploymentModification     string
	ResetDeploymentModification      string
	DeleteDeployment                 string
	SelectDeployments                string

	InsertPlugin  string
	SelectPlugin  string
	DeletePlugin  string
	SelectPlugins string

	CreateTemplates                    string
	CreateTemplatesMetadata            string
	CreateTemplatesMetadataIndex       string
	CreateTemplatesDeployments         string
	CreateSites                        string
	CreateSitesMetadata                string
	CreateSitesMetadataIndex           string
	CreateSitesDeployments             string
	CreateDeployments                  string
	CreateDeploymentsPreparedIndex     string
	CreateDeploymentsModificationIndex string
	CreatePlugins                      string

	DropTemplates                    string
	DropTemplatesMetadata            string
	DropTemplatesMetadataIndex       string
	DropTemplatesDeployments         string
	DropSites                        string
	DropSitesMetadata                string
	DropSitesMetadataIndex           string
	DropSitesDeployments             string
	DropDeployments                  string
	DropDeploymentsPreparedIndex     string
	DropDeploymentsModificationIndex string
	DropPlugins                      string

	PreparedSelectTemplate              *sql.Stmt
	PreparedDeleteTemplate              *sql.Stmt
	PreparedSelectSite                  *sql.Stmt
	PreparedDeleteSite                  *sql.Stmt
	PreparedSelectDeployment            *sql.Stmt
	PreparedDeleteDeployment            *sql.Stmt
	PreparedResetDeploymentModification *sql.Stmt
	PreparedInsertPlugin                *sql.Stmt
	PreparedSelectPlugin                *sql.Stmt
	PreparedDeletePlugin                *sql.Stmt
	PreparedSelectPlugins               *sql.Stmt

	db  *sql.DB
	log commonlog.Logger
}

func NewStatements(driver string, db *sql.DB, log commonlog.Logger) *Statements {
	switch driver {
	case "pgx":
		return NewPostgresqlStatements(db, log)

	default:
		panic(fmt.Sprintf("unsupported SQL driver: %s", driver))
	}
}

func (self *Statements) Prepare(context contextpkg.Context) error {
	stmts := []**sql.Stmt{
		&self.PreparedSelectTemplate,
		&self.PreparedDeleteTemplate,
		&self.PreparedSelectSite,
		&self.PreparedDeleteSite,
		&self.PreparedSelectDeployment,
		&self.PreparedDeleteDeployment,
		&self.PreparedResetDeploymentModification,
		&self.PreparedInsertPlugin,
		&self.PreparedSelectPlugin,
		&self.PreparedDeletePlugin,
		&self.PreparedSelectPlugins,
	}

	statements := []string{
		self.SelectTemplate,
		self.DeleteTemplate,
		self.SelectSite,
		self.DeleteSite,
		self.SelectDeployment,
		self.DeleteDeployment,
		self.ResetDeploymentModification,
		self.InsertPlugin,
		self.SelectPlugin,
		self.DeletePlugin,
		self.SelectPlugins,
	}

	for index, stmt := range stmts {
		var err error
		if *stmt, err = self.db.PrepareContext(context, statements[index]); err != nil {
			self.log.Critical(statements[index])
			return err
		}
	}

	return nil
}

func (self *Statements) Release() {
	stmts := []*sql.Stmt{
		self.PreparedSelectTemplate,
		self.PreparedDeleteTemplate,
		self.PreparedSelectSite,
		self.PreparedDeleteSite,
		self.PreparedSelectDeployment,
		self.PreparedDeleteDeployment,
		self.PreparedResetDeploymentModification,
		self.PreparedInsertPlugin,
		self.PreparedSelectPlugin,
		self.PreparedDeletePlugin,
		self.PreparedSelectPlugins,
	}

	for _, stmt := range stmts {
		if stmt != nil {
			if err := stmt.Close(); err != nil {
				self.log.Error(err.Error())
			}
		}
	}
}

func (self *Statements) CreateTables(context contextpkg.Context) error {
	return self.execAll(context,
		self.CreateTemplates,
		self.CreateTemplatesMetadata,
		self.CreateTemplatesMetadataIndex,
		self.CreateSites,
		self.CreateSitesMetadata,
		self.CreateSitesMetadataIndex,
		self.CreateDeployments,
		self.CreateDeploymentsPreparedIndex,
		self.CreateDeploymentsModificationIndex,
		self.CreateTemplatesDeployments,
		self.CreateSitesDeployments,
		self.CreatePlugins,
	)
}

func (self *Statements) DropTables(context contextpkg.Context) error {
	return self.execAll(context,
		self.DropSitesDeployments,
		self.DropTemplatesDeployments,
		self.DropPlugins,
		self.DropDeploymentsModificationIndex,
		self.DropDeploymentsPreparedIndex,
		self.DropDeployments,
		self.DropSitesMetadataIndex,
		self.DropSitesMetadata,
		self.DropSites,
		self.DropTemplatesMetadataIndex,
		self.DropTemplatesMetadata,
		self.DropTemplates,
	)
}

// Utils

func (self *Statements) execAll(context contextpkg.Context, statements ...string) error {
	for _, statement := range statements {
		if _, err := self.db.ExecContext(context, statement); err != nil {
			self.log.Critical(statement)
			return err
		}
	}
	return nil
}

func nilIfEmptyString(s string) any {
	if s == "" {
		return nil
	} else {
		return s
	}
}
