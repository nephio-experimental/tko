package metascheduling

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
)

type SchedulerFunc func(context contextpkg.Context, schedulingContext *Context) error

func (self *MetaScheduling) RegisterScheduler(gvk util.GVK, schedule SchedulerFunc) {
	self.schedulers[gvk] = schedule
}

func (self *MetaScheduling) GetScheduler(gvk util.GVK) (SchedulerFunc, bool, error) {
	if schedule, ok := self.schedulers[gvk]; ok {
		return schedule, true, nil
	} else if plugin, ok, err := self.Client.GetPlugin(client.NewPluginID("schedule", gvk)); err == nil {
		if ok {
			if schedule, err := NewPluginScheduler(plugin); err == nil {
				return schedule, true, nil
			} else {
				return nil, false, err
			}
		}
	} else {
		return nil, false, err
	}
	return nil, false, nil
}
