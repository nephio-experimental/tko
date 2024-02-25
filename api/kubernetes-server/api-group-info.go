package server

import (
	krmgroup "github.com/nephio-experimental/tko/api/krm/tko.nephio.org"
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	apiserver "k8s.io/apiserver/pkg/server"
)

func NewAPIGroupInfo(restOptions generic.RESTOptionsGetter, backend backend.Backend, log commonlog.Logger) (*apiserver.APIGroupInfo, error) {
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

	return &apiGroupInfo, nil
}
