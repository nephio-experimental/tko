package scheduling

import (
	"sync"
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// Scheduling
//

type Scheduling struct {
	Client  *clientpkg.Client
	Timeout time.Duration
	Log     commonlog.Logger

	registeredSchedulers map[util.GVK][]ScheduleFunc
	schedulers           sync.Map
}

func NewScheduling(client *clientpkg.Client, timeout time.Duration, log commonlog.Logger) *Scheduling {
	self := Scheduling{
		Client:               client,
		Log:                  log,
		registeredSchedulers: make(map[util.GVK][]ScheduleFunc),
	}

	return &self
}

func (self *Scheduling) ResetPluginCache() {
	self.schedulers = sync.Map{}
}
