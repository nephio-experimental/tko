package v1alpha1

import (
	krmgroup "github.com/nephio-experimental/tko/api/krm/tko.nephio.org"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Group-Version.
// Note: kube_codegen *requires* it to be named "SchemeGroupVersion".
var SchemeGroupVersion = schema.GroupVersion{Group: krmgroup.GroupName, Version: Version}

// Kind to Group-Kind.
// Note: kube_codegen *requires* it to be called "Kind".
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource name to Group-Resource. For both singular and plural resource names.
// Note: kube_codegen *requires* it to be named "Resource".
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Adds all our types to a scheme.
// Note: kube_codegen *requires* it to be named "AddToScheme".
var AddToScheme = schemeBuilder.AddToScheme

var schemeBuilder = runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		new(Template),
		new(TemplateList),
		new(Site),
		new(SiteList),
		new(Deployment),
		new(DeploymentList),
		new(Plugin),
		new(PluginList),
	)
	meta.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
})

// Internal Group-Version.
var InternalSchemeGroupVersion = schema.GroupVersion{Group: krmgroup.GroupName, Version: runtime.APIVersionInternal}

// Adds all our internally-versioned types to a scheme.
var AddInternalToScheme = internalSchemeBuilder.AddToScheme

var internalSchemeBuilder = runtime.NewSchemeBuilder(func(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(InternalSchemeGroupVersion,
		new(Template),
		new(TemplateList),
		new(Site),
		new(SiteList),
		new(Deployment),
		new(DeploymentList),
		new(Plugin),
		new(PluginList),
	)
	return nil
})
