package util

import (
	"github.com/tliron/go-ard"
)

func MergeResources(resources Resources, mergeResources Resources) Resources {
	for _, mergeResource := range mergeResources {
		if resourceIdentifier, ok := NewResourceIdentifierForResource(mergeResource); ok {
			renameAnnotation, _ := GetRenameAnnotation(mergeResource)
			if renameAnnotation != "" {
				resourceIdentifier.Name = renameAnnotation
				mergeResource = ard.Copy(mergeResource).(Resource)
				ard.With(mergeResource).Get("metadata", "name").Set(renameAnnotation)
			}

			add := true

			for index, resource := range resources {
				if resourceIdentifier.Is(resource) {
					mergeAnnotation, _ := GetMergeAnnotation(mergeResource)
					switch mergeAnnotation {
					case MergeAnnotationReplace, "":
						resources[index] = mergeResource
					case MergeAnnotationOverride:
						ard.Merge(resources[index], mergeResource, false)
					}

					add = false
					break
				}
			}

			if add {
				resources = append(resources, mergeResource)
			}
		}
	}
	return resources
}

func PrepareResourcesForMerge(resources Resources) Resources {
	resources_ := make(Resources, len(resources))
	for index, resource := range resources {
		resource = ard.Copy(resource).(Resource)
		resources_[index] = resource
		UpdateAnnotationsForMerge(resource)
	}
	return resources_
}
