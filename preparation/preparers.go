package preparation

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

type PreparerFunc func(context contextpkg.Context, preparationContext *Context) (bool, []ard.Map, error)

func (self *Preparation) RegisterPreparer(gvk tkoutil.GVK, prepare PreparerFunc) {
	self.preparers[gvk] = prepare
}

var prepareString = "prepare"

func (self *Preparation) GetPreparers(gvk tkoutil.GVK) ([]PreparerFunc, error) {
	var preparers []PreparerFunc

	if prepare, ok := self.preparers[gvk]; ok {
		preparers = append(preparers, prepare)
	}

	if plugins, err := self.Client.ListPlugins(client.ListPlugins{
		Type:    &prepareString,
		Trigger: &gvk,
	}); err == nil {
		if util.IterateResults(plugins, func(plugin client.Plugin) error {
			if prepare, err := NewPluginPreparer(plugin); err == nil {
				preparers = append(preparers, prepare)
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

	return preparers, nil
}

func (self *Preparation) ShouldPrepare(resourceIdentifier tkoutil.ResourceIdentifier, resource tkoutil.Resource, log commonlog.Logger) (bool, []PreparerFunc) {
	if prepareAnnotation, ok := tkoutil.GetPrepareAnnotation(resource); ok {
		if prepareAnnotation == tkoutil.PrepareAnnotationHere {
			if preparers, err := self.GetPreparers(resourceIdentifier.GVK); err == nil {
				if len(preparers) == 0 {
					if log != nil {
						log.Error("no preparers registered for trigger",
							"resourceType", resourceIdentifier.GVK)
					}
					return false, nil
				}

				if !tkoutil.IsPreparedAnnotation(resource) {
					return true, preparers
				} else if log != nil {
					log.Info("already prepared",
						"resource", resourceIdentifier)
				}
			} else if log != nil {
				log.Error(err.Error())
			}
		}
	} else if preparers, err := self.GetPreparers(resourceIdentifier.GVK); err == nil {
		// If there is no annotation but there *is* at least one preparer then we will still try to prepare (as if the annotation is "Here")
		if len(preparers) > 0 {
			if !tkoutil.IsPreparedAnnotation(resource) {
				return true, preparers
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
