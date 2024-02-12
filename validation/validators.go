package validation

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
)

type ValidatorFunc func(context contextpkg.Context, validationContext *Context) []error

func (self *Validation) RegisterValidator(gvk util.GVK, validate ValidatorFunc) {
	self.validators[gvk] = validate
}

func (self *Validation) GetValidator(gvk util.GVK) (ValidatorFunc, bool, error) {
	if validator, ok := self.validators[gvk]; ok {
		return validator, true, nil
	} else if plugin, ok, err := self.Client.GetPlugin(client.NewPluginID("validate", gvk)); err == nil {
		if ok {
			if validate, err := NewPluginValidator(plugin); err == nil {
				return validate, true, nil
			} else {
				return nil, false, err
			}
		}
	} else {
		return nil, false, err
	}
	return nil, false, nil
}
