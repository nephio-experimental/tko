package sql

import (
	contextpkg "context"
	"database/sql"
	"fmt"

	"github.com/nephio-experimental/tko/backend"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

var _ backend.Backend = new(SQLBackend)

//
// SQLBackend
//

type SQLBackend struct {
	DropTablesFirst bool

	driver          string
	dataSource      string
	resourcesFormat string

	statements *Statements
	db         *sql.DB

	log                     commonlog.Logger
	maxModificationDuration int64 // microseconds
}

func NewSQLBackend(driver string, dataSource string, resourcesFormat string, maxModificationDurationSeconds float64, log commonlog.Logger) *SQLBackend {
	return &SQLBackend{
		driver:                  driver,
		dataSource:              dataSource,
		resourcesFormat:         resourcesFormat,
		log:                     log,
		maxModificationDuration: int64(maxModificationDurationSeconds * 1_000_000.0),
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) Connect(context contextpkg.Context) error {
	self.log.Noticef("connect: driver=%s dataSource=%s", self.driver, self.dataSource)
	var err error
	if self.db, err = sql.Open(self.driver, self.dataSource); err == nil {
		self.statements = NewStatements(self.driver, self.db, self.log)

		if self.DropTablesFirst {
			err = self.statements.DropTables(context)
			if err != nil {
				return err
			}
		}

		err = self.statements.CreateTables(context)
		if err != nil {
			return err
		}

		err = self.statements.Prepare(context)
		if err != nil {
			return err
		}

		return nil
	} else {
		return err
	}
}

// ([backend.Backend] interface)
func (self *SQLBackend) Release(context contextpkg.Context) error {
	self.log.Noticef("release: driver=%s dataSource=%s", self.driver, self.dataSource)
	if self.statements != nil {
		self.statements.Release()
	}
	if self.db != nil {
		return self.db.Close()
	} else {
		return nil
	}
}

// ([fmt.Stringer] interface)
// ([backend.Backend] interface)
func (self *SQLBackend) String() string {
	return fmt.Sprintf("SQL driver=%s dataSource=%s resourcesFormat=%s", self.driver, self.dataSource, self.resourcesFormat)
}

// Utils

func (self *SQLBackend) rollback(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		self.log.Error(err.Error())
	}
}

func (self *SQLBackend) closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		self.log.Error(err.Error())
	}
}

func (self *SQLBackend) encodeResources(resources util.Resources) ([]byte, error) {
	return util.EncodeResources(self.resourcesFormat, resources)
}

func (self *SQLBackend) decodeResources(content []byte) (util.Resources, error) {
	return util.DecodeResources(self.resourcesFormat, content)
}
