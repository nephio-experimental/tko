package v1alpha1

// Note: kube_codegen *requires* this file to be named "types.go".
// Also make sure JSON names are lower-camel-case versions of Go names.

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//
// Template
//

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
	// +optional
	TemplateId *string `json:"templateId"`
	// +optional
	Metadata map[string]string `json:"metadata"`
	// +optional
	DeploymentIds []string `json:"deploymentIds,omitempty"`
	// +optional
	ResourcesYaml *string `json:"resourcesYaml"`
}

type TemplateStatus struct {
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
	// +optional
	SiteId *string `json:"siteId"`
	// +optional
	TemplateId *string `json:"templateId,omitempty"`
	// +optional
	Metadata map[string]string `json:"metadata"`
	// +optional
	DeploymentIds []string `json:"deploymentIds,omitempty"`
	// +optional
	ResourcesYaml *string `json:"resourcesYaml"`
}

type SiteStatus struct {
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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Deployment struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSpec   `json:"spec"`
	Status DeploymentStatus `json:"status,omitempty"`
}

type DeploymentSpec struct {
	// +optional
	DeploymentId *string `json:"deploymentId"`
	// +optional
	ParentDeploymentId *string `json:"parentDeploymentId,omitempty"`
	// +optional
	TemplateId *string `json:"templateId,omitempty"`
	// +optional
	SiteId *string `json:"siteId,omitempty"`
	// +optional
	Metadata map[string]string `json:"metadata"`
	// +optional
	Prepared *bool `json:"prepared"`
	// +optional
	Approved *bool `json:"approved"`
	// +optional
	ResourcesYaml *string `json:"resourcesYaml"`
}

type DeploymentStatus struct {
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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Plugin struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   PluginSpec   `json:"spec"`
	Status PluginStatus `json:"status,omitempty"`
}

type PluginSpec struct {
	Type *string `json:"type"`
	// +optional
	PluginID *string `json:"id"`
	Executor *string `json:"executor"`
	// +optional
	Arguments []string `json:"arguments"`
	// +optional
	Properties map[string]string `json:"properties,omitempty"`
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
