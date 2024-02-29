package server

import (
	krmgroup "github.com/nephio-experimental/tko/api/krm/tko.nephio.org"
	krm "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/server"
	apiserver "k8s.io/apiserver/pkg/server"
	serverpkg "k8s.io/apiserver/pkg/server"

	// Support *all* authentication methods (increases the size of the executable)
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

//
// Server
//

type Server struct {
	Backend backend.Backend
	Log     commonlog.Logger

	apiServer *server.GenericAPIServer
	stop      chan struct{}
	stopped   chan struct{}
}

func NewServer(backend backend.Backend, port int, log commonlog.Logger) (*Server, error) {
	server := Server{
		Backend: backend,
		Log:     log,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}

	if config, err := NewConfig(port); err == nil {
		if server.apiServer, err = config.New(ServerName, serverpkg.NewEmptyDelegate()); err == nil {
			if err := server.apiServer.InstallAPIGroup(NewAPIGroupInfo(config.RESTOptionsGetter, backend, log)); err == nil {
				return &server, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Server) Start() {
	self.Log.Notice("starting Kubernetes server")
	go func() {
		if err := self.apiServer.PrepareRun().Run(self.stop); err != nil {
			self.Log.Error(err.Error())
		}
		self.stopped <- struct{}{}
	}()
}

func (self *Server) Stop() {
	self.Log.Notice("stopping Kubernetes server")
	self.stop <- struct{}{}
	<-self.stopped
	self.Log.Notice("stopped Kubernetes server")
}

func NewAPIGroupInfo(restOptions generic.RESTOptionsGetter, backend backend.Backend, log commonlog.Logger) *apiserver.APIGroupInfo {
	templateStore := NewTemplateStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "template"))
	siteStore := NewSiteStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "site"))
	deploymentStore := NewDeploymentStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "deployment"))
	pluginStore := NewPluginStore(backend, commonlog.NewKeyValueLogger(log, "resourceType", "plugin"))

	apiGroupInfo := apiserver.NewDefaultAPIGroupInfo(krmgroup.GroupName, Scheme, meta.ParameterCodec, Codecs)
	apiGroupInfo.VersionedResourcesStorageMap[krm.Version] = map[string]rest.Storage{
		templateStore.TypePlural:   templateStore,
		siteStore.TypePlural:       siteStore,
		deploymentStore.TypePlural: deploymentStore,
		pluginStore.TypePlural:     pluginStore,
	}

	return &apiGroupInfo
}
