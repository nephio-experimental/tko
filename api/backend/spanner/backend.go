package spanner

import (
	"context"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"github.com/tliron/commonlog"
)

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
func (self *SpannerBackend) Connect() error {
	var err error
	self.admin, err = database.NewDatabaseAdminClient(context.TODO())
	if err != nil {
		return err
	}

	self.client, err = spanner.NewClient(context.TODO(), self.path)
	if err != nil {
		self.admin.Close()
		self.admin = nil
		return err
	}

	return nil
}

// ([Backend] interface)
func (self *SpannerBackend) Release() error {
	if self.client != nil {
		self.client.Close()
	}
	if self.admin != nil {
		return self.admin.Close()
	}
	return nil
}
