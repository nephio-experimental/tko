package scheduling

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type ScheduleFunc func(context contextpkg.Context, schedulingContext *Context) error

type Schedulers []ScheduleFunc

type SchedulersMap map[tkoutil.GVK]Schedulers

func (self *Scheduling) RegisterScheduler(gvk tkoutil.GVK, schedule ScheduleFunc) {
	schedulers, _ := self.registeredSchedulers[gvk]
	schedulers = append(schedulers, schedule)
	self.registeredSchedulers[gvk] = schedulers
}

var scheduleString = "schedule"

func (self *Scheduling) GetSchedulers(gvk tkoutil.GVK) (Schedulers, error) {
	if schedulers, ok := self.schedulers.Load(gvk); ok {
		return schedulers.(Schedulers), nil
	}

	var schedulers Schedulers

	if schedulers_, ok := self.registeredSchedulers[gvk]; ok {
		schedulers = append(schedulers, schedulers_...)
	}

	if plugins, err := self.Client.ListPlugins(client.ListPlugins{
		Type:    &scheduleString,
		Trigger: &gvk,
	}); err == nil {
		if err := util.IterateResults(plugins, func(plugin client.Plugin) error {
			if schedule, err := NewPluginScheduler(plugin, self.LogIPStack, self.LogAddress, self.LogPort); err == nil {
				schedulers = append(schedulers, schedule)
				return nil
			} else {
				return err
			}
		}); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	if schedulers_, loaded := self.schedulers.LoadOrStore(gvk, schedulers); loaded {
		schedulers = schedulers_.(Schedulers)
	}

	return schedulers, nil
}
