package sql

import (
	contextpkg "context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/reflection"
)

//
// Statements
//

type Statements struct {
	// Templates

	CreateTemplates              string
	DropTemplates                string
	CreateTemplatesMetadata      string
	DropTemplatesMetadata        string
	CreateTemplatesMetadataIndex string
	DropTemplatesMetadataIndex   string
	CreateTemplatesDeployments   string
	DropTemplatesDeployments     string

	UpsertTemplate           string
	UpsertTemplateMetadata   string
	UpsertTemplateDeployment string
	SelectTemplate           string
	DeleteTemplate           string
	DeleteTemplateMetadata   string
	DeleteTemplateDeployment string
	SelectTemplates          string

	// Sites

	CreateSites              string
	DropSites                string
	CreateSitesMetadata      string
	DropSitesMetadata        string
	CreateSitesMetadataIndex string
	DropSitesMetadataIndex   string
	CreateSitesDeployments   string
	DropSitesDeployments     string

	UpsertSite           string
	UpsertSiteMetadata   string
	UpsertSiteDeployment string
	SelectSite           string
	DeleteSite           string
	DeleteSiteMetadata   string
	DeleteSiteDeployment string
	SelectSites          string

	// Deployments

	CreateDeployments                  string
	DropDeployments                    string
	CreateDeploymentsMetadata          string
	DropDeploymentsMetadata            string
	CreateDeploymentsMetadataIndex     string
	DropDeploymentsMetadataIndex       string
	CreateDeploymentsPreparedIndex     string
	DropDeploymentsPreparedIndex       string
	CreateDeploymentsApprovedIndex     string
	DropDeploymentsApprovedIndex       string
	CreateDeploymentsModificationIndex string
	DropDeploymentsModificationIndex   string

	InsertDeployment                 string
	UpdateDeployment                 string
	UpsertDeploymentMetadata         string
	SelectDeployment                 string
	SelectDeploymentWithModification string
	SelectDeploymentByModification   string
	UpdateDeploymentModification     string
	ResetDeploymentModification      string
	DeleteDeployment                 string
	DeleteDeploymentMetadata         string
	SelectDeployments                string

	// Plugins

	CreatePlugins         string
	DropPlugins           string
	CreatePluginsTriggers string
	DropPluginsTriggers   string

	UpsertPlugin         string
	InsertPluginTrigger  string
	SelectPlugin         string
	DeletePlugin         string
	DeletePluginTriggers string
	SelectPlugins        string

	// These statements will be automatically prepared and released
	// The source SQL field has the same name without the "Prepared" prefix

	PreparedUpsertTemplate                        *sql.Stmt
	PreparedUpsertTemplateMetadata                *sql.Stmt
	PreparedUpsertTemplateDeployment              *sql.Stmt
	PreparedSelectTemplate                        *sql.Stmt
	PreparedDeleteTemplate                        *sql.Stmt
	PreparedDeleteTemplateMetadata                *sql.Stmt
	PreparedDeleteTemplateDeployment              *sql.Stmt
	PreparedDeleteTemplateDeploymentsByDeployment *sql.Stmt
	PreparedUpsertSite                            *sql.Stmt
	PreparedUpsertSiteMetadata                    *sql.Stmt
	PreparedUpsertSiteDeployment                  *sql.Stmt
	PreparedSelectSite                            *sql.Stmt
	PreparedDeleteSite                            *sql.Stmt
	PreparedDeleteSiteMetadata                    *sql.Stmt
	PreparedDeleteSiteDeployment                  *sql.Stmt
	PreparedInsertDeployment                      *sql.Stmt
	PreparedUpdateDeployment                      *sql.Stmt
	PreparedUpsertDeploymentMetadata              *sql.Stmt
	PreparedSelectDeployment                      *sql.Stmt
	PreparedSelectDeploymentWithModification      *sql.Stmt
	PreparedSelectDeploymentByModification        *sql.Stmt
	PreparedUpdateDeploymentModification          *sql.Stmt
	PreparedResetDeploymentModification           *sql.Stmt
	PreparedDeleteDeployment                      *sql.Stmt
	PreparedDeleteDeploymentMetadata              *sql.Stmt
	PreparedUpsertPlugin                          *sql.Stmt
	PreparedInsertPluginTrigger                   *sql.Stmt
	PreparedSelectPlugin                          *sql.Stmt
	PreparedDeletePlugin                          *sql.Stmt
	PreparedDeletePluginTriggers                  *sql.Stmt

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
	for _, preparedStmtField := range preparedStmtFields {
		self.log.Debugf("preparing statement: %s", preparedStmtField.SourceName)
		if stmt, err := self.db.PrepareContext(context, preparedStmtField.GetSource(self)); err == nil {
			preparedStmtField.SetPrepared(self, stmt)
		} else {
			return fmt.Errorf("preparing %s: %w", preparedStmtField.SourceName, err)
		}
	}

	return nil
}

func (self *Statements) Release() {
	for _, preparedStmtField := range preparedStmtFields {
		if stmt := preparedStmtField.GetPrepared(self); stmt != nil {
			self.log.Debugf("releasing statement: %s", preparedStmtField.SourceName)
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
		self.CreateDeploymentsMetadata,
		self.CreateDeploymentsMetadataIndex,
		self.CreateDeploymentsPreparedIndex,
		self.CreateDeploymentsApprovedIndex,
		self.CreateDeploymentsModificationIndex,

		self.CreateTemplatesDeployments,
		self.CreateSitesDeployments,

		self.CreatePlugins,
		self.CreatePluginsTriggers,
	)
}

func (self *Statements) DropTables(context contextpkg.Context) error {
	return self.execAll(context,
		self.DropPluginsTriggers,
		self.DropPlugins,

		self.DropSitesDeployments,
		self.DropTemplatesDeployments,

		self.DropDeploymentsMetadataIndex,
		self.DropDeploymentsMetadata,
		self.DropDeploymentsModificationIndex,
		self.DropDeploymentsPreparedIndex,
		self.DropDeploymentsApprovedIndex,
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

//
// PreparedStmtField
//

type PreparedStmtField struct {
	SourceName   string
	PreparedName string
}

func (self PreparedStmtField) GetSource(statements *Statements) string {
	statements_ := reflect.ValueOf(statements).Elem()
	return statements_.FieldByName(self.SourceName).Interface().(string) // will panic if not string
}

func (self PreparedStmtField) GetPrepared(statements *Statements) *sql.Stmt {
	statements_ := reflect.ValueOf(statements).Elem()
	return statements_.FieldByName(self.SourceName).Interface().(*sql.Stmt) // will panic if not *sql.Stmt
}

func (self PreparedStmtField) SetPrepared(statements *Statements, stmt *sql.Stmt) {
	statements_ := reflect.ValueOf(statements).Elem()
	statements_.FieldByName(self.PreparedName).Set(reflect.ValueOf(stmt)) // will panic if not *sql.Stmt
}

const preparedPrefix = "Prepared"

var preparedStmtFields []PreparedStmtField

func init() {
	stmtType := reflect.TypeFor[*sql.Stmt]()
	preparedPrefixLength := len(preparedPrefix)
	structFields := reflection.GetStructFields(reflect.TypeFor[Statements]())
	for _, structField := range structFields {
		if (structField.Type == stmtType) && strings.HasPrefix(structField.Name, preparedPrefix) {
			sourceName := structField.Name[preparedPrefixLength:]
			for _, structField_ := range structFields {
				if structField_.Name == sourceName {
					preparedStmtFields = append(preparedStmtFields, PreparedStmtField{
						SourceName:   sourceName,
						PreparedName: structField.Name,
					})
				}
			}
		}
	}
}
