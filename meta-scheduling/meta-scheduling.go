package metascheduling

import (
	"sync"
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

	registeredSchedulers map[util.GVK][]SchedulerFunc
	schedulers           sync.Map
}

func NewMetaScheduling(client *clientpkg.Client, timeout time.Duration, log commonlog.Logger) *MetaScheduling {
	return &MetaScheduling{
		Client:               client,
		Log:                  log,
		registeredSchedulers: make(map[util.GVK][]SchedulerFunc),
	}
}
