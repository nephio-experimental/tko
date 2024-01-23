package metascheduling

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
)

type SchedulerFunc func(schedulingContext *Context) error

func (self *MetaScheduling) RegisterScheduler(gvk util.GVK, scheduler SchedulerFunc) {
	self.schedulers[gvk] = scheduler
}

func (self *MetaScheduling) GetScheduler(gvk util.GVK) (SchedulerFunc, bool, error) {
	if scheduler, ok := self.schedulers[gvk]; ok {
		return scheduler, true, nil
	} else if plugin, ok, err := self.Client.GetPlugin(client.NewPluginID("schedule", gvk)); err == nil {
		if ok {
			if scheduler, err := NewPluginScheduler(plugin); err == nil {
				return scheduler, true, nil
			} else {
				return nil, false, err
			}
		}
	} else {
		return nil, false, err
	}
	return nil, false, nil
}
