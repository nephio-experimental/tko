package validation

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
)

type ValidatorFunc func(context *Context) error

func (self *Validation) RegisterValidator(gvk util.GVK, instantiator ValidatorFunc) {
	self.validators[gvk] = instantiator
}

func (self *Validation) GetValidator(gvk util.GVK) (ValidatorFunc, bool, error) {
	if validator, ok := self.validators[gvk]; ok {
		return validator, true, nil
	} else if plugin, ok, err := self.Client.GetPlugin(client.NewPluginID("validate", gvk)); err == nil {
		if ok {
			if validator, err := NewPluginValidator(plugin); err == nil {
				return validator, true, nil
			} else {
				return nil, false, err
			}
		}
	} else {
		return nil, false, err
	}
	return nil, false, nil
}
