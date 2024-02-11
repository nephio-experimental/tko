package metascheduling

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

	MetaScheduling *MetaScheduling

	log commonlog.Logger
}

func NewController(metaScheduling *MetaScheduling, interval time.Duration, log commonlog.Logger) *Controller {
	self := Controller{
		MetaScheduling: metaScheduling,
		log:            log,
	}
	self.Controller = util.NewController(self.run, interval, log)
	return &self
}

func (self *Controller) run() error {
	if err := self.MetaScheduling.ScheduleSites(); err != nil {
		self.log.Error(err.Error())
	}
	return nil
}
