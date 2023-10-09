package preparation

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

	Preparation *Preparation

	log commonlog.Logger
}

func NewController(preparation *Preparation, log commonlog.Logger) *Controller {
	self := Controller{
		Preparation: preparation,
		log:         log,
	}
	self.Controller = util.NewController(self.run, 3*time.Second, log)
	return &self
}

func (self *Controller) run() error {
	if err := self.Preparation.PrepareDeployments(); err != nil {
		self.log.Error(err.Error())
	}
	return nil
}
