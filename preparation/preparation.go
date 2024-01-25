package preparation

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/nephio-experimental/tko/validation"
	"github.com/tliron/commonlog"
)

//
// Preparation
//

type Preparation struct {
	Client     *client.Client
	Validation *validation.Validation
	Log        commonlog.Logger

	preparers map[util.GVK]PreparerFunc
}

func NewPreparation(client_ *client.Client, validation *validation.Validation, log commonlog.Logger) *Preparation {
	return &Preparation{
		Client:     client_,
		Validation: validation,
		Log:        log,
		preparers:  make(map[util.GVK]PreparerFunc),
	}
}
