package util

import (
	"fmt"

	"github.com/tliron/go-ard"
)

type Resource = ard.Map

type Resources = []Resource

var DeploymentGVK = NewGVK("deployment.nephio.org", "v1alpha1", "Deployment")

var DeploymentResourceIdentifier = DeploymentGVK.NewResourceIdentifier("deployment")

func CloneResources(resources Resources) Resources {
	return ard.Copy(resources).(Resources)
}

func GetReferentResources(objectReferences ard.List, resources Resources) (Resources, error) {
	var referentResources Resources
	for _, objectReference := range objectReferences {
		if objectReference_, ok := objectReference.(ard.Map); ok {
			if resourceIdentifier, ok := NewResourceIdentifierForObjectReference(objectReference_); ok {
				if resource, ok := resourceIdentifier.GetResource(resources); ok {
					referentResources = append(referentResources, resource)
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
	return referentResources, nil
}

func NewDeploymentResource(templateId string, siteId string, prepared bool, approved bool) Resource {
	spec := ard.Map{
		"templateId": templateId,
	}

	if siteId != "" {
		spec["siteId"] = siteId
	}

	if prepared {
		spec["prepared"] = true
	}

	if approved {
		spec["approved"] = true
	}

	return Resource{
		"apiVersion": DeploymentResourceIdentifier.GVK.APIVersion(),
		"kind":       DeploymentResourceIdentifier.GVK.Kind,
		"metadata": ard.Map{
			"name": DeploymentResourceIdentifier.Name,
		},
		"spec": spec,
	}
}
