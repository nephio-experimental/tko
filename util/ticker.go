package util

import (
	"time"
)

//
// Ticker
//

type Ticker struct {
	F func()

	duration time.Duration
	ticker   *time.Ticker
	stop     chan struct{}
}

func NewTicker(duration time.Duration, f func()) *Ticker {
	return &Ticker{
		F:        f,
		duration: duration,
		stop:     make(chan struct{}),
	}
}

func (self *Ticker) Start() {
	ticker := time.NewTicker(self.duration)
	self.ticker = ticker

	go func() {
		for {
			select {
			case <-self.stop:
				return
			case <-ticker.C:
				self.F()
			}
		}
	}()
}

func (self *Ticker) Stop() {
	if self.ticker != nil {
		self.stop <- struct{}{}
		self.ticker.Stop()
		self.ticker = nil
	}
}
