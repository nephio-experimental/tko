package topology

import (
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

var TemplateGVK = util.NewGVK("topology.nephio.org", "v1alpha1", "Template")

func GetTemplateID(resources []util.Resource, name string) (string, bool) {
	if template, ok := TemplateGVK.NewResourceIdentifier(name).GetResource(resources); ok {
		if templateId, ok := ard.With(template).Get("spec", "explicit", "id").String(); ok {
			return templateId, true
		}

		// TODO: support implicit matching
	}
	return "", false
}
