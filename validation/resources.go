package validation

import (
	contextpkg "context"
	"errors"

	"github.com/nephio-experimental/tko/util"
)

func (self *Validation) ValidateResources(resources util.Resources, complete bool) error {
	var errs []error

	for _, resource := range resources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if validate, ok, err := self.GetValidator(resourceIdentifier.GVK); err == nil {
				if !ok {
					validate = self.DefaultValidate
				}

				validationContext := self.NewContext(resources, resourceIdentifier, complete)
				context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
				errs = append(errs, validate(context, validationContext)...)
				cancel()
			} else {
				errs = append(errs, err)
			}
		} else {
			errs = append(errs, self.Kubeconform(resource, complete)...)
		}
	}

	return errors.Join(errs...)
}

// ([ValidatorFunc] signature)
func (self *Validation) DefaultValidate(context contextpkg.Context, validationContext *Context) []error {
	if resource, ok := validationContext.GetResource(); ok {
		return self.Kubeconform(resource, validationContext.Complete)
	} else {
		return nil
	}
}
