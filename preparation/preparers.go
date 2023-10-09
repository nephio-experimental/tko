package preparation

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-ard"
)

type PreparerFunc func(context *Context) (bool, []ard.Map, error)

func (self *Preparation) RegisterPreparer(gvk util.GVK, preparer PreparerFunc) {
	self.preparers[gvk] = preparer
}

func (self *Preparation) GetPreparer(gvk util.GVK) (PreparerFunc, bool, error) {
	if preparer, ok := self.preparers[gvk]; ok {
		return preparer, true, nil
	} else if plugin, ok, err := self.Client.GetPlugin(client.NewPluginID("prepare", gvk)); err == nil {
		if ok {
			if preparer, err := NewPluginPreparer(plugin); err == nil {
				return preparer, true, nil
			} else {
				return nil, false, err
			}
		}
	} else {
		return nil, false, err
	}
	return nil, false, nil
}

func (self *Preparation) ShouldPrepare(resourceIdentifier util.ResourceIdentifier, resource util.Resource, log commonlog.Logger) (bool, PreparerFunc) {
	if prepareAnnotation, ok := util.GetPrepareAnnotation(resource); ok {
		if prepareAnnotation == util.PrepareAnnotationHere {
			if preparer, ok, err := self.GetPreparer(resourceIdentifier.GVK); err == nil {
				if ok {
					if !util.IsPreparedAnnotation(resource) {
						return true, preparer
					} else if log != nil {
						log.Infof("already prepared: %s", resourceIdentifier)
					}
				} else {
					if log != nil {
						log.Errorf("plugin not registered: %s", resourceIdentifier)
					}
					return true, nil
				}
			} else if log != nil {
				log.Error(err.Error())
			}
		}
	} else if preparer, ok, err := self.GetPreparer(resourceIdentifier.GVK); err == nil {
		// If there is no annotation but there *is* a preparer then we will still try to prepare (as if the annotation is "Here")
		if ok {
			if !util.IsPreparedAnnotation(resource) {
				return true, preparer
			} else if log != nil {
				log.Infof("already prepared: %s", resourceIdentifier)
			}
		}
	} else if log != nil {
		log.Error(err.Error())
	}
	return false, nil
}
