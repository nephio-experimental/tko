package scheduling

import (
	"sync"
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

//
// Scheduling
//

type Scheduling struct {
	Client     *clientpkg.Client
	Timeout    time.Duration
	Log        commonlog.Logger
	LogIPStack util.IPStack
	LogAddress string
	LogPort    int

	registeredSchedulers SchedulersMap
	schedulers           sync.Map
}

func NewScheduling(client *clientpkg.Client, timeout time.Duration, log commonlog.Logger, logIpStack util.IPStack, logAddress string, logPort int) *Scheduling {
	self := Scheduling{
		Client:               client,
		Timeout:              timeout,
		LogIPStack:           logIpStack,
		LogAddress:           logAddress,
		LogPort:              logPort,
		Log:                  log,
		registeredSchedulers: make(SchedulersMap),
	}

	return &self
}

func (self *Scheduling) ResetPluginCache() {
	self.schedulers = sync.Map{}
}
