package metascheduling

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type SchedulerFunc func(context contextpkg.Context, schedulingContext *Context) error

func (self *MetaScheduling) RegisterScheduler(gvk tkoutil.GVK, schedule SchedulerFunc) {
	schedulers, _ := self.registeredSchedulers[gvk]
	schedulers = append(schedulers, schedule)
	self.registeredSchedulers[gvk] = schedulers
}

var scheduleString = "schedule"

func (self *MetaScheduling) GetSchedulers(gvk tkoutil.GVK) ([]SchedulerFunc, error) {
	if schedulers, ok := self.schedulers.Load(gvk); ok {
		return schedulers.([]SchedulerFunc), nil
	}

	var schedulers []SchedulerFunc

	if schedulers_, ok := self.registeredSchedulers[gvk]; ok {
		schedulers = append(schedulers, schedulers_...)
	}

	if plugins, err := self.Client.ListPlugins(client.ListPlugins{
		Type:    &scheduleString,
		Trigger: &gvk,
	}); err == nil {
		if util.IterateResults(plugins, func(plugin client.Plugin) error {
			if schedule, err := NewPluginScheduler(plugin); err == nil {
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
		schedulers = schedulers_.([]SchedulerFunc)
	}

	return schedulers, nil
}
