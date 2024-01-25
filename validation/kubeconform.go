package validation

import (
	"bytes"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-transcribe"
	resourcepkg "github.com/yannh/kubeconform/pkg/resource"
	validatorpkg "github.com/yannh/kubeconform/pkg/validator"
)

func (self *Validation) Kubeconform(resource util.Resource, complete bool) []error {
	if !complete {
		// Kubeconform does not support partial validation
		return nil
	}

	var errs []error

	var buffer bytes.Buffer
	if err := transcribe.NewTranscriber().SetWriter(&buffer).SetIndentSpaces(2).WriteYAML(resource); err == nil {
		resource_ := resourcepkg.Resource{Path: "unknown", Bytes: buffer.Bytes()}
		result := self.kubeconform.ValidateResource(resource_)
		if result.Err != nil {
			errs = append(errs, result.Err)
		} else if result.Status == validatorpkg.Invalid {
			for _, err := range result.ValidationErrors {
				errs = append(errs, &err)
			}
		}
	}

	return errs
}
