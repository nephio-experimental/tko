package scheduling

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

	Scheduling *Scheduling

	log commonlog.Logger
}

func NewController(scheduling *Scheduling, interval time.Duration, log commonlog.Logger) *Controller {
	self := Controller{
		Scheduling: scheduling,
		log:        log,
	}
	self.Controller = util.NewController(self.run, interval, log)
	return &self
}

func (self *Controller) run() error {
	if err := self.Scheduling.ScheduleSites(); err != nil {
		self.log.Error(err.Error())
	}
	return nil
}
