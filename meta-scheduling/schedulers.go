package metascheduling

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type SchedulerFunc func(context contextpkg.Context, schedulingContext *Context) error

func (self *MetaScheduling) RegisterScheduler(gvk tkoutil.GVK, schedule SchedulerFunc) {
	self.schedulers[gvk] = schedule
}

var scheduleString = "schedule"

func (self *MetaScheduling) GetSchedulers(gvk tkoutil.GVK) ([]SchedulerFunc, error) {
	var schedulers []SchedulerFunc

	if schedule, ok := self.schedulers[gvk]; ok {
		schedulers = append(schedulers, schedule)
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

	return schedulers, nil
}
