package server

import (
	krmgroup "github.com/nephio-experimental/tko/api/krm/tko.nephio.org"
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	apiserver "k8s.io/apiserver/pkg/server"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)

func init() {
	runtimeutil.Must(krm.AddToScheme(Scheme))

	// See: https://github.com/kubernetes/sample-apiserver/blob/bd85aa5bdde49be0c9aa611678f08059bd2a248f/pkg/apiserver/apiserver.go#L46
	v1 := schema.GroupVersion{Version: "v1"}
	Scheme.AddUnversionedTypes(v1,
		new(meta.Status),
		new(meta.APIVersions),
		new(meta.APIGroupList),
		new(meta.APIGroup),
		new(meta.APIResourceList),
	)
	meta.AddToGroupVersion(Scheme, v1)
}

func NewAPIGroupInfo(restOptions generic.RESTOptionsGetter, backend backend.Backend, log commonlog.Logger) *apiserver.APIGroupInfo {
	templateStore := NewTemplateStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "template"))
	siteStore := NewSiteStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "site"))
	deploymentStore := NewDeploymentStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "deployment"))
	pluginStore := NewPluginStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "plugin"))

	apiGroupInfo := apiserver.NewDefaultAPIGroupInfo(krmgroup.GroupName, Scheme, meta.ParameterCodec, Codecs)
	apiGroupInfo.VersionedResourcesStorageMap[krm.Version] = map[string]rest.Storage{
		templateStore.Plural:   templateStore,
		siteStore.Plural:       siteStore,
		deploymentStore.Plural: deploymentStore,
		pluginStore.Plural:     pluginStore,
	}

	return &apiGroupInfo
}
