package validation

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type ValidatorFunc func(context contextpkg.Context, validationContext *Context) []error

func (self *Validation) RegisterValidator(gvk tkoutil.GVK, validate ValidatorFunc) {
	validators, _ := self.registeredValidators[gvk]
	validators = append(validators, validate)
	self.registeredValidators[gvk] = validators
}

var validateString = "validate"

func (self *Validation) GetValidators(gvk tkoutil.GVK, complete bool) ([]ValidatorFunc, error) {
	if validators, ok := self.validators.Load(gvk); ok {
		return self.defaultValidators(validators.([]ValidatorFunc), complete), nil
	}

	var validators []ValidatorFunc

	if validators_, ok := self.registeredValidators[gvk]; ok {
		validators = append(validators, validators_...)
	}

	if plugins, err := self.Client.ListPlugins(client.ListPlugins{
		Type:    &validateString,
		Trigger: &gvk,
	}); err == nil {
		if util.IterateResults(plugins, func(plugin client.Plugin) error {
			if validate, err := NewPluginValidator(plugin); err == nil {
				validators = append(validators, validate)
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

	if validators_, loaded := self.validators.LoadOrStore(gvk, validators); loaded {
		validators = validators_.([]ValidatorFunc)
	}

	return self.defaultValidators(validators, complete), nil
}

func (self *Validation) defaultValidators(validators []ValidatorFunc, complete bool) []ValidatorFunc {
	if complete && (len(validators) == 0) {
		validators = []ValidatorFunc{self.DefaultValidate}
	}
	return validators
}
