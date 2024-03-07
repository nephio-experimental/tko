package preparation

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

type PrepareFunc func(context contextpkg.Context, preparationContext *Context) (bool, []ard.Map, error)

type Preparers []PrepareFunc

type PreparersMap map[tkoutil.GVK]Preparers

func (self *Preparation) RegisterPreparer(gvk tkoutil.GVK, prepare PrepareFunc) {
	preparers, _ := self.registeredPreparers[gvk]
	preparers = append(preparers, prepare)
	self.registeredPreparers[gvk] = preparers
}

var prepareString = "prepare"

// TODO: cache
func (self *Preparation) GetPreparers(gvk tkoutil.GVK) (Preparers, error) {
	if preparers, ok := self.preparers.Load(gvk); ok {
		return preparers.(Preparers), nil
	}

	var preparers Preparers

	if preparers_, ok := self.registeredPreparers[gvk]; ok {
		preparers = append(preparers, preparers_...)
	}

	if plugins, err := self.Client.ListPlugins(client.ListPlugins{
		Type:    &prepareString,
		Trigger: &gvk,
	}); err == nil {
		if err := util.IterateResults(plugins, func(plugin client.Plugin) error {
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

	if preparers_, loaded := self.preparers.LoadOrStore(gvk, preparers); loaded {
		preparers = preparers_.(Preparers)
	}

	return preparers, nil
}
