package validation

import (
	"bytes"
	"errors"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-transcribe"
	resourcepkg "github.com/yannh/kubeconform/pkg/resource"
	validatorpkg "github.com/yannh/kubeconform/pkg/validator"
)

func (self *Validation) ValidateResources(resources util.Resources, complete bool) error {
	var errs []error

	for _, resource := range resources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if validator, ok, err := self.GetValidator(resourceIdentifier.GVK); err == nil {
				if ok {
					context := self.NewContext(resources, resourceIdentifier, complete)
					if err := validator(context); err != nil {
						errs = append(errs, err)
					}
				} else {
					if complete {
						errs = append(errs, self.DefaultValidateCompleteResource(resource)...)
					}
				}
			} else {
				errs = append(errs, err)
			}
		} else {
			if complete {
				errs = append(errs, self.DefaultValidateCompleteResource(resource)...)
			}
		}
	}

	return errors.Join(errs...)
}

func (self *Validation) DefaultValidateCompleteResource(resource util.Resource) []error {
	var errs []error

	var buffer bytes.Buffer
	if err := (&transcribe.Transcriber{Writer: &buffer, Indent: "  "}).WriteYAML(resource); err == nil {
		resource_ := resourcepkg.Resource{Path: "unknown", Bytes: buffer.Bytes()}
		result := self.kubeconform.ValidateResource(resource_)
		if result.Err != nil {
			errs = append(errs, result.Err)
		} else if result.Status == validatorpkg.Invalid {
			for _, e := range result.ValidationErrors {
				errs = append(errs, &e)
			}
		}
	}

	return errs
}
