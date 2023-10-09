package util

import (
	contextpkg "context"
	"time"

	"github.com/tliron/commonlog"
)

//
// Controller
//

type Controller struct {
	Run      func() error
	Interval time.Duration
	Log      commonlog.Logger

	context contextpkg.Context
	stop    contextpkg.CancelFunc
	stopped chan struct{}
}

func NewController(run func() error, interval time.Duration, log commonlog.Logger) *Controller {
	context, stop := contextpkg.WithCancel(contextpkg.Background())
	return &Controller{
		Run:      run,
		Interval: interval,
		Log:      log,
		context:  context,
		stop:     stop,
		stopped:  make(chan struct{}),
	}
}

func (self *Controller) Start() {
	self.Log.Notice("starting controller")
	go func() {
		for {
			select {
			case <-time.After(self.Interval):
				if err := self.Run(); err != nil {
					self.Log.Criticalf("stopped controller due to error: %s", err.Error())
					self.stopped <- struct{}{}
					return
				}

			case <-self.context.Done():
				self.Log.Notice("stopped controller")
				self.stopped <- struct{}{}
				return
			}
		}
	}()
}

func (self *Controller) Stop() {
	self.Log.Notice("stopping controller")
	self.stop()
	<-self.stopped
}
