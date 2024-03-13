package topology

import (
	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/preparation"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

var TemplateGVK = tkoutil.NewGVK("topology.nephio.org", "v1alpha1", "Template")

// TODO: cache result
func GetTemplateID(preparationContext *preparation.Context, name string) (string, bool) {
	if template, ok := TemplateGVK.NewResourceIdentifier(name).GetResource(preparationContext.DeploymentPackage); ok {
		spec := ard.With(template).Get("spec").ConvertSimilar()

		if templateId, ok := spec.Get("templateId").String(); ok {
			return templateId, true
		}

		if selectMetadata, ok := spec.Get("select", "metadata").StringMap(); ok {
			metadataPatterns := make(map[string]string)
			for key, value := range selectMetadata {
				metadataPatterns[key] = util.ToString(value)
			}

			if templateInfos, err := preparationContext.Preparation.Client.ListTemplates(clientpkg.SelectTemplates{MetadataPatterns: metadataPatterns}, 0, 1); err == nil {
				defer templateInfos.Release()

				// First one we find
				if templateInfo, err := templateInfos.Next(); err == nil {
					return templateInfo.TemplateID, true
				} else {
					preparationContext.Log.Error(err.Error())
				}
			}
		}
	}

	return "", false
}
