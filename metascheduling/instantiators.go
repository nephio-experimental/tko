package metascheduling

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
)

type InstantiatorFunc func(instantiationContext *Context) error

func (self *Instantiation) RegisterInstantiator(gvk util.GVK, instantiator InstantiatorFunc) {
	self.instantiators[gvk] = instantiator
}

func (self *Instantiation) GetInstantiator(gvk util.GVK) (InstantiatorFunc, bool, error) {
	if instantiator, ok := self.instantiators[gvk]; ok {
		return instantiator, true, nil
	} else if plugin, ok, err := self.Client.GetPlugin(client.NewPluginID("instantiate", gvk)); err == nil {
		if ok {
			if instantiator, err := NewPluginInstantiator(plugin); err == nil {
				return instantiator, true, nil
			} else {
				return nil, false, err
			}
		}
	} else {
		return nil, false, err
	}
	return nil, false, nil
}
