package instantiation

import (
	"time"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// Controller
//

type Controller struct {
	*util.Controller

	Instantiation *Instantiation

	log commonlog.Logger
}

func NewController(instantiation *Instantiation, log commonlog.Logger) *Controller {
	self := Controller{
		Instantiation: instantiation,
		log:           log,
	}
	self.Controller = util.NewController(self.run, 3*time.Second, log)
	return &self
}

func (self *Controller) run() error {
	if err := self.Instantiation.InstantiateSites(); err != nil {
		self.log.Error(err.Error())
	}
	return nil
}
