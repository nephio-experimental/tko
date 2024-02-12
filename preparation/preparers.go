package preparation

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-ard"
)

type PreparerFunc func(context contextpkg.Context, preparationContext *Context) (bool, []ard.Map, error)

func (self *Preparation) RegisterPreparer(gvk util.GVK, prepare PreparerFunc) {
	self.preparers[gvk] = prepare
}

func (self *Preparation) GetPreparer(gvk util.GVK) (PreparerFunc, bool, error) {
	if prepare, ok := self.preparers[gvk]; ok {
		return prepare, true, nil
	} else if plugin, ok, err := self.Client.GetPlugin(client.NewPluginID("prepare", gvk)); err == nil {
		if ok {
			if prepare, err := NewPluginPreparer(plugin); err == nil {
				return prepare, true, nil
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
			if prepare, ok, err := self.GetPreparer(resourceIdentifier.GVK); err == nil {
				if ok {
					if !util.IsPreparedAnnotation(resource) {
						return true, prepare
					} else if log != nil {
						log.Info("already prepared",
							"resource", resourceIdentifier)
					}
				} else {
					if log != nil {
						log.Error("plugin not registered",
							"resourceType", resourceIdentifier.GVK)
					}
					return true, nil
				}
			} else if log != nil {
				log.Error(err.Error())
			}
		}
	} else if prepare, ok, err := self.GetPreparer(resourceIdentifier.GVK); err == nil {
		// If there is no annotation but there *is* a preparer then we will still try to prepare (as if the annotation is "Here")
		if ok {
			if !util.IsPreparedAnnotation(resource) {
				return true, prepare
			} else if log != nil {
				log.Info("already prepared",
					"resource", resourceIdentifier)
			}
		}
	} else if log != nil {
		log.Error(err.Error())
	}
	return false, nil
}
