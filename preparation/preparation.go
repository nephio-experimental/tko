package preparation

import (
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/nephio-experimental/tko/validation"
	"github.com/tliron/commonlog"
)

//
// Preparation
//

type Preparation struct {
	Client     *clientpkg.Client
	Validation *validation.Validation
	Timeout    time.Duration
	Log        commonlog.Logger

	preparers map[util.GVK]PreparerFunc
}

func NewPreparation(client *clientpkg.Client, validation *validation.Validation, timeout time.Duration, log commonlog.Logger) *Preparation {
	return &Preparation{
		Client:     client,
		Validation: validation,
		Timeout:    timeout,
		Log:        log,
		preparers:  make(map[util.GVK]PreparerFunc),
	}
}
