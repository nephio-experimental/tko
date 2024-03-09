package preparation

import (
	"sync"
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

//
// Preparation
//

type Preparation struct {
	Client      *clientpkg.Client
	Timeout     time.Duration
	AutoApprove bool
	Log         commonlog.Logger
	LogIPStack  util.IPStack
	LogAddress  string
	LogPort     int

	registeredPreparers PreparersMap
	preparers           sync.Map
}

func NewPreparation(client *clientpkg.Client, timeout time.Duration, autoApprove bool, log commonlog.Logger, logIpStack util.IPStack, logAddress string, logPort int) *Preparation {
	return &Preparation{
		Client:              client,
		Timeout:             timeout,
		AutoApprove:         autoApprove,
		Log:                 log,
		LogIPStack:          logIpStack,
		LogAddress:          logAddress,
		LogPort:             logPort,
		registeredPreparers: make(PreparersMap),
	}
}

func (self *Preparation) ResetPluginCache() {
	self.preparers = sync.Map{}
}
