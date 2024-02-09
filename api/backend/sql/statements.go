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
	InsertTemplate           string
	InsertTemplateMetadata   string
	InsertTemplateDeployment string
	SelectTemplate           string
	DeleteTemplate           string
	SelectTemplates          string

	InsertSite           string
	InsertSiteMetadata   string
	InsertSiteDeployment string
	SelectSite           string
	DeleteSite           string
	SelectSites          string

	InsertDeployment                 string
	InsertDeploymentMetadata         string
	UpdateDeployment                 string
	SelectDeployment                 string
	SelectDeploymentWithModification string
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
	CreateDeploymentsMetadata          string
	CreateDeploymentsMetadataIndex     string
	CreateDeploymentsPreparedIndex     string
	CreateDeploymentsApprovedIndex     string
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
	DropDeploymentsMetadata          string
	DropDeploymentsMetadataIndex     string
	DropDeploymentsPreparedIndex     string
	DropDeploymentsApprovedIndex     string
	DropDeploymentsModificationIndex string
	DropPlugins                      string

	// These statements will be automatically prepared and released
	// The source SQL field has the same name without the "Prepared" prefix
	PreparedInsertTemplate                   *sql.Stmt
	PreparedInsertTemplateMetadata           *sql.Stmt
	PreparedInsertTemplateDeployment         *sql.Stmt
	PreparedSelectTemplate                   *sql.Stmt
	PreparedDeleteTemplate                   *sql.Stmt
	PreparedInsertSite                       *sql.Stmt
	PreparedInsertSiteMetadata               *sql.Stmt
	PreparedInsertSiteDeployment             *sql.Stmt
	PreparedSelectSite                       *sql.Stmt
	PreparedDeleteSite                       *sql.Stmt
	PreparedInsertDeployment                 *sql.Stmt
	PreparedInsertDeploymentMetadata         *sql.Stmt
	PreparedUpdateDeployment                 *sql.Stmt
	PreparedSelectDeployment                 *sql.Stmt
	PreparedSelectDeploymentWithModification *sql.Stmt
	PreparedSelectDeploymentByModification   *sql.Stmt
	PreparedUpdateDeploymentModification     *sql.Stmt
	PreparedResetDeploymentModification      *sql.Stmt
	PreparedDeleteDeployment                 *sql.Stmt
	PreparedInsertPlugin                     *sql.Stmt
	PreparedSelectPlugin                     *sql.Stmt
	PreparedDeletePlugin                     *sql.Stmt
	PreparedSelectPlugins                    *sql.Stmt

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
			return err
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
	)
}

func (self *Statements) DropTables(context contextpkg.Context) error {
	return self.execAll(context,
		self.DropSitesDeployments,
		self.DropTemplatesDeployments,
		self.DropPlugins,
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
