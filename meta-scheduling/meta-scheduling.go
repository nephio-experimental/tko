package metascheduling

import (
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// MetaScheduling
//

type MetaScheduling struct {
	Client  *clientpkg.Client
	Timeout time.Duration
	Log     commonlog.Logger

	schedulers map[util.GVK]SchedulerFunc
}

func NewMetaScheduling(client *clientpkg.Client, timeout time.Duration, log commonlog.Logger) *MetaScheduling {
	return &MetaScheduling{
		Client:     client,
		Log:        log,
		schedulers: make(map[util.GVK]SchedulerFunc),
	}
}
