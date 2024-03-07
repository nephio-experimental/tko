package scheduling

import (
	"sync"
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/tliron/commonlog"
)

//
// Scheduling
//

type Scheduling struct {
	Client  *clientpkg.Client
	Timeout time.Duration
	Log     commonlog.Logger

	registeredSchedulers SchedulersMap
	schedulers           sync.Map
}

func NewScheduling(client *clientpkg.Client, timeout time.Duration, log commonlog.Logger) *Scheduling {
	self := Scheduling{
		Client:               client,
		Timeout:              timeout,
		Log:                  log,
		registeredSchedulers: make(SchedulersMap),
	}

	return &self
}

func (self *Scheduling) ResetPluginCache() {
	self.schedulers = sync.Map{}
}
