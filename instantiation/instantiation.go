package instantiation

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// Instantiation
//

type Instantiation struct {
	Client *client.Client
	Log    commonlog.Logger

	instantiators map[util.GVK]InstantiatorFunc
}

func NewInstantiation(client_ *client.Client, log commonlog.Logger) *Instantiation {
	return &Instantiation{
		Client:        client_,
		Log:           log,
		instantiators: make(map[util.GVK]InstantiatorFunc),
	}
}
