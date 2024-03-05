package util

import (
	"fmt"

	"github.com/tliron/go-ard"
)

//
// Package
//

type Package = []Resource

func ClonePackage(package_ Package) Package {
	return ard.Copy(package_).(Package)
}

func GetReferentPackage(objectReferences ard.List, package_ Package) (Package, error) {
	var referentPackage Package
	for _, objectReference := range objectReferences {
		if objectReference_, ok := objectReference.(ard.Map); ok {
			if resourceIdentifier, ok := NewResourceIdentifierForObjectReference(objectReference_); ok {
				if resource, ok := resourceIdentifier.GetResource(package_); ok {
					referentPackage = append(referentPackage, resource)
				} else {
					return nil, fmt.Errorf("object reference not found: %s", resourceIdentifier)
				}
			} else {
				return nil, fmt.Errorf("malformed object reference: %s", objectReference_)
			}
		} else {
			return nil, fmt.Errorf("object reference not a map: %s", objectReference)
		}
	}
	return referentPackage, nil
}

func MergePackage(package_ Package, mergePackage ...Resource) Package {
	for _, mergeResource := range mergePackage {
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

			for index, resource := range package_ {
				if mergeResourceIdentifier.Is(resource) {
					if override {
						package_[index] = ard.Merge(resource, mergeResource, false).(Resource)
					} else {
						package_[index] = mergeResource
					}

					add = false
					break
				}
			}

			if add {
				package_ = append(package_, mergeResource)
			}
		}
	}
	return package_
}

func PreparePackageForMerge(package_ Package) Package {
	package__ := make(Package, len(package_))
	for index, resource := range package_ {
		resource = ard.Copy(resource).(Resource)
		package__[index] = resource
		UpdateAnnotationsForMerge(resource)
	}
	return package__
}
