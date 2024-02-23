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
	TemplateId *string           `json:"templateId"`
	Metadata   map[string]string `json:"metadata"`
}

type TemplateStatus struct {
	Test string `json:"test"`
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
	SiteId   *string           `json:"siteId"`
	Metadata map[string]string `json:"metadata"`
}

type SiteStatus struct {
	Test string `json:"test"`
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
	Test string `json:"test"`
}

type DeploymentStatus struct {
	Test string `json:"test"`
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
	Test string `json:"test"`
}

type PluginStatus struct {
	Test string `json:"test"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PluginList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`

	Items []Plugin `json:"items"`
}
