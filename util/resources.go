package util

import (
	"github.com/tliron/go-ard"
)

//
// ResourceIdentifier
//

type ResourceIdentifier struct {
	GVK  GVK    `json:",inline" yaml:",inline"`
	Name string `json:"name" yaml:"name"`
}

func (self GVK) NewResourceIdentifier(name string) ResourceIdentifier {
	return ResourceIdentifier{self, name}
}

func NewResourceIdentifierForResource(resource Resource) (ResourceIdentifier, bool) {
	var self ResourceIdentifier
	var ok bool
	if self.GVK, ok = GetGVK(resource); ok {
		if self.Name, ok = ard.With(resource).ConvertSimilar().Get("metadata", "name").String(); ok {
			return self, true
		}
	}
	return self, false
}

func NewResourceIdentifierForObjectReference(objectReference ard.Map) (ResourceIdentifier, bool) {
	var self ResourceIdentifier
	var ok bool
	if self.GVK, ok = GetGVK(objectReference); ok {
		if self.Name, ok = ard.With(objectReference).ConvertSimilar().Get("name").String(); ok {
			return self, true
		}
	}
	return self, false
}

func (self ResourceIdentifier) GetResource(package_ Package) (Resource, bool) {
	for _, resource := range package_ {
		if self.Is(resource) {
			return resource, true
		}
	}
	return nil, false
}

func (self ResourceIdentifier) Is(resource Resource) bool {
	if self.GVK.Is(resource) {
		if name, ok := ard.With(resource).ConvertSimilar().Get("metadata", "name").String(); ok {
			return name == self.Name
		}
	}
	return false
}

// ([fmt.Stringer] interface)
func (self ResourceIdentifier) String() string {
	return self.GVK.String() + ", name: " + self.Name
}

//
// ResourceIdentifiers
//

type ResourceIdentifiers struct {
	list []ResourceIdentifier
}

func (self *ResourceIdentifiers) Empty() bool {
	return len(self.list) == 0
}

func (self *ResourceIdentifiers) Push(resourceIdentifier ResourceIdentifier) {
	self.list = append(self.list, resourceIdentifier)
}

func (self *ResourceIdentifiers) Pop() (ResourceIdentifier, bool) {
	if !self.Empty() {
		todo := self.list[0]
		self.list = self.list[1:]
		return todo, true
	} else {
		return ResourceIdentifier{}, false
	}
}

//
// Resource
//

type Resource = ard.Map

var DeploymentGVK = NewGVK("deployment.nephio.org", "v1alpha1", "Deployment")

var DeploymentResourceIdentifier = DeploymentGVK.NewResourceIdentifier("deployment")

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
			"annotations": ard.Map{
				MergeAnnotation: MergeAnnotationOverride,
			},
		},
		"spec": spec,
	}
}
