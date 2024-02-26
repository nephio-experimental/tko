package v1alpha1

// Note: kube_codegen *requires* this file to be named "types.go".
// Also make sure JSON names are lower-camel-case versions of Go names
// and that all properties are JSON-marshallable.

import (
	"github.com/tliron/go-ard"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	openapi "k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

//
// Template
//

// TKO template.
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Template struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec"`
	Status TemplateStatus `json:"status,omitempty"`
}

type TemplateSpec struct {
	// Template ID. Must be unique per TKO instance.
	// +optional
	TemplateId *string `json:"templateId"`

	// Template metadata.
	// +optional
	Metadata map[string]string `json:"metadata"`

	// Template KRM package. The KRM must be at least *partially* valid,
	// meaning that required properties at any level of nesting are allowed
	// to be missing, but properties that are assigned must have valid
	// values.
	// +optional
	Package *Package `json:"package"`
}

type TemplateStatus struct {
	// (Read only) IDs of deployments created from this template.
	// +optional
	DeploymentIds []string `json:"deploymentIds,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TemplateList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Template `json:"items"`
}

//
// Site
//

// TKO site.
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Site struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   SiteSpec   `json:"spec"`
	Status SiteStatus `json:"status,omitempty"`
}

type SiteSpec struct {
	// Site ID. Must be unique per TKO instance.
	// +optional
	SiteId *string `json:"siteId"`

	// ID of the template from which this site was created.
	// +optional
	TemplateId *string `json:"templateId,omitempty"`

	// Site metadata.
	// +optional
	Metadata map[string]string `json:"metadata"`

	// Site KRM package. The KRM must be completely valid.
	// +optional
	Package *Package `json:"package"`
}

type SiteStatus struct {
	// (Read only) IDs of deployments associated with this site.
	// +optional
	DeploymentIds []string `json:"deploymentIds,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SiteList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Site `json:"items"`
}

//
// Deployment
//

// TKO deployment.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Deployment struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSpec   `json:"spec"`
	Status DeploymentStatus `json:"status,omitempty"`
}

type DeploymentSpec struct {
	// (Read only) Deployment ID. A random UUID generated when the
	// deployment is created.
	// +optional
	DeploymentId *string `json:"deploymentId"`

	// ID of the deployment that created this deployment during the
	// preparation process.
	// +optional
	ParentDeploymentId *string `json:"parentDeploymentId,omitempty"`

	// ID of the template from which this deployment was created.
	// +optional
	TemplateId *string `json:"templateId,omitempty"`

	// ID of the site with which this deployment is associated.
	// +optional
	SiteId *string `json:"siteId,omitempty"`

	// Deployment metadata.
	// +optional
	Metadata map[string]string `json:"metadata"`

	// Deployment KRM package. When "prepared" is true the KRM must be
	// completely valid. Otherwise, the KRM must be at least *partially*
	// valid, meaning that required properties at any level of nesting are
	// allowed to be missing, but properties that are assigned must have
	// valid values.
	// +optional
	Package *Package `json:"package"`
}

type DeploymentStatus struct {
	// True if this deployment is prepared.
	// +optional
	Prepared *bool `json:"prepared"`

	// True if this deployment is approved. Cannot be true if
	// "prepared" is false.
	// +optional
	Approved *bool `json:"approved"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DeploymentList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Deployment `json:"items"`
}

//
// Plugin
//

// TKO plugin.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Plugin struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   PluginSpec   `json:"spec"`
	Status PluginStatus `json:"status,omitempty"`
}

type PluginSpec struct {
	// Can be "validate", "prepare", or "schedule".
	Type *string `json:"type"`

	// Plugin ID. Must be unique per "type" per TKO instance.
	// +optional
	PluginID *string `json:"id"`

	// Plugin executor. Each executor has its own requirements for
	// "arguments" and "properties".
	Executor *string `json:"executor"`

	// Sequence of arguments for the executor.
	// +optional
	Arguments []string `json:"arguments"`

	// Map of properties for the executor.
	// +optional
	Properties map[string]string `json:"properties,omitempty"`

	// List of KRM types (GVK) that trigger this plugin.
	// +optional
	Triggers []Trigger `json:"triggers,omitempty"`
}

type Trigger struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type PluginStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PluginList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Plugin `json:"items"`
}

//
// Package
//

// KRM package.
type Package struct {
	// KRM package contents.
	Resources []ard.StringMap `json:"resources,omitempty"`
}

func (self *Package) DeepCopyInto(out *Package) {
	resources := make([]ard.StringMap, len(self.Resources))
	for index, resource := range self.Resources {
		resources[index] = ard.Copy(resource).(ard.StringMap)
	}
	out.Resources = resources
}

func (self *Package) DeepCopy() *Package {
	resources := new(Package)
	self.DeepCopyInto(resources)
	return resources
}

func (_ Package) OpenAPIDefinition() openapi.OpenAPIDefinition {
	return openapi.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "KRM package.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"resources": {
						SchemaProps: spec.SchemaProps{
							Description: "KRM package contents.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Description: "Resource in the KRM package.",
										Type:        []string{"object"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
