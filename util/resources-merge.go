package util

import (
	"github.com/tliron/go-ard"
)

func MergeResources(resources Resources, mergeResources ...Resource) Resources {
	for _, mergeResource := range mergeResources {
		if mergeResourceIdentifier, ok := NewResourceIdentifierForResource(mergeResource); ok {
			var override bool
			if mergeAnnotation, ok := GetMergeAnnotation(mergeResource); ok {
				if mergeAnnotation == MergeAnnotationOverride {
					override = true
				}
			}

			if renameAnnotation, ok := GetRenameAnnotation(mergeResource); ok {
				mergeResourceIdentifier.Name = renameAnnotation
				mergeResource = ard.Copy(mergeResource).(Resource)
				ard.With(mergeResource).ConvertSimilar().ForceGet("metadata", "name").Set(renameAnnotation)
			}

			add := true

			for index, resource := range resources {
				if mergeResourceIdentifier.Is(resource) {
					if override {
						resources[index] = ard.Merge(resource, mergeResource, false).(Resource)
					} else {
						resources[index] = mergeResource
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
