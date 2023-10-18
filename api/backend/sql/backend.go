package sql

import (
	"database/sql"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// SqlBackend
//

type SqlBackend struct {
	driver          string
	dataSource      string
	resourcesFormat string

	sql *Sql
	db  *sql.DB

	log                commonlog.Logger
	modificationWindow int64 // microseconds
}

// modificationWindow in seconds
func NewSqlBackend(driver string, dataSource string, resourcesFormat string, modificationWindow int, log commonlog.Logger) *SqlBackend {
	return &SqlBackend{
		driver:             driver,
		dataSource:         dataSource,
		resourcesFormat:    resourcesFormat,
		log:                log,
		modificationWindow: int64(modificationWindow) * 1000000,
	}
}

// ([backend.Backend] interface)
func (self *SqlBackend) Connect() error {
	self.log.Notice("connect")
	var err error
	if self.db, err = sql.Open(self.driver, self.dataSource); err == nil {
		self.sql = NewSql(self.driver, self.db, self.log)

		err = self.sql.DropTables()
		if err != nil {
			return err
		}

		err = self.sql.CreateTables()
		if err != nil {
			return err
		}

		err = self.sql.Prepare()
		if err != nil {
			return err
		}

		return nil
	} else {
		return err
	}
}

// ([backend.Backend] interface)
func (self *SqlBackend) Release() error {
	self.log.Notice("release")
	if self.sql != nil {
		self.sql.Release()
	}
	if self.db != nil {
		return self.db.Close()
	} else {
		return nil
	}
}

// Utils

func (self *SqlBackend) encodeResources(resources util.Resources) ([]byte, error) {
	return util.EncodeResources(self.resourcesFormat, resources)
}

func (self *SqlBackend) decodeResources(content []byte) (util.Resources, error) {
	return util.DecodeResources(self.resourcesFormat, content)
}
