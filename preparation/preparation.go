package preparation

import (
	"sync"
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// Preparation
//

type Preparation struct {
	Client      *clientpkg.Client
	Timeout     time.Duration
	AutoApprove bool
	Log         commonlog.Logger

	registeredPreparers map[util.GVK][]PrepareFunc
	preparers           sync.Map
}

func NewPreparation(client *clientpkg.Client, timeout time.Duration, autoApprove bool, log commonlog.Logger) *Preparation {
	return &Preparation{
		Client:              client,
		Timeout:             timeout,
		AutoApprove:         autoApprove,
		Log:                 log,
		registeredPreparers: make(map[util.GVK][]PrepareFunc),
	}
}

func (self *Preparation) ResetPluginCache() {
	self.preparers = sync.Map{}
}
