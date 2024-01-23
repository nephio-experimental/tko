package metascheduling

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// MetaScheduling
//

type MetaScheduling struct {
	Client *client.Client
	Log    commonlog.Logger

	schedulers map[util.GVK]SchedulerFunc
}

func NewMetaScheduling(client_ *client.Client, log commonlog.Logger) *MetaScheduling {
	return &MetaScheduling{
		Client:     client_,
		Log:        log,
		schedulers: make(map[util.GVK]SchedulerFunc),
	}
}
