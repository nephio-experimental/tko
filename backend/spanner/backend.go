package spanner

import (
	contextpkg "context"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
)

var _ backend.Backend = new(SpannerBackend)

//
// SpannerBackend
//

type SpannerBackend struct {
	path string

	admin  *database.DatabaseAdminClient
	client *spanner.Client

	log commonlog.Logger
}

func NewSpannerBackend(path string, log commonlog.Logger) *SpannerBackend {
	return &SpannerBackend{
		path: path,
		log:  log,
	}
}

// ([Backend] interface)
func (self *SpannerBackend) Connect(context contextpkg.Context) error {
	var err error
	self.admin, err = database.NewDatabaseAdminClient(context)
	if err != nil {
		return err
	}

	self.client, err = spanner.NewClient(context, self.path)
	if err != nil {
		self.admin.Close()
		self.admin = nil
		return err
	}

	return nil
}

// ([Backend] interface)
func (self *SpannerBackend) Release(context contextpkg.Context) error {
	if self.client != nil {
		self.client.Close()
	}
	if self.admin != nil {
		return self.admin.Close()
	}
	return nil
}

// ([fmt.Stringer] interface)
// ([backend.Backend] interface)
func (self *SpannerBackend) String() string {
	return "Spanner"
}
