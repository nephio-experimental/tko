package backend

import (
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

func cloneMetadata(metadata map[string]string) map[string]string {
	metadata_ := make(map[string]string)
	for key, value := range metadata {
		metadata_[key] = value
	}
	return metadata_
}

func cloneResources(resources util.Resources) util.Resources {
	return ard.Copy(resources).(util.Resources)
}

func updateMetadata(metadata map[string]string, resources util.Resources) {
	for _, resource := range resources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if resourceIdentifier.GVK.Group == "metadata.nephio.org" {
				if annotation, ok := util.GetMetadataAnnotation(resource); ok {
					switch annotation {
					case util.MetadataAnnotationHere, "":
					case util.MetadataAnnotationPostpone, util.MetadataAnnotationNever:
						continue
					}
				}

				if spec, ok := ard.With(resource).Get("spec").Map(); ok {
					updateMetadataValues(metadata, resourceIdentifier.GVK.Kind+".", spec)
				}
			}
		}
	}
}

func updateMetadataValues(metadata map[string]string, prefix string, values ard.Map) {
	for key, value := range values {
		key_ := ard.MapKeyToString(key)
		switch value_ := value.(type) {
		case ard.Map:
			updateMetadataValues(metadata, prefix+key_+".", value_)

		default:
			metadata[prefix+key_] = ard.ValueToString(value)
		}
	}
}
