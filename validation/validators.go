package validation

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type ValidatorFunc func(context contextpkg.Context, validationContext *Context) []error

func (self *Validation) RegisterValidator(gvk tkoutil.GVK, validate ValidatorFunc) {
	self.validators[gvk] = validate
}

var validateString = "validate"

func (self *Validation) GetValidators(gvk tkoutil.GVK, complete bool) ([]ValidatorFunc, error) {
	var validators []ValidatorFunc

	if validate, ok := self.validators[gvk]; ok {
		validators = append(validators, validate)
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

	if complete && (len(validators) == 0) {
		return []ValidatorFunc{self.DefaultValidate}, nil
	}

	return validators, nil
}
