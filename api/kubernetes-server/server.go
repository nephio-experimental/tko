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
	Port    int
	Log     commonlog.Logger

	apiServer *server.GenericAPIServer
	stop      chan struct{}
	stopped   chan struct{}
}

func NewServer(backend backend.Backend, port int, log commonlog.Logger) *Server {
	return &Server{
		Backend: backend,
		Port:    port,
		Log:     log,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (self *Server) Start() error {
	self.Log.Notice("starting Kubernetes server")

	if config, err := NewConfig(self.Port); err == nil {
		if self.apiServer, err = config.New(ServerName, serverpkg.NewEmptyDelegate()); err == nil {
			if err := self.apiServer.InstallAPIGroup(NewAPIGroupInfo(config.RESTOptionsGetter, self.Backend, self.Log)); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}

	go func() {
		if err := self.apiServer.PrepareRun().Run(self.stop); err != nil {
			self.Log.Error(err.Error())
		}
		self.stopped <- struct{}{}
	}()

	return nil
}

func (self *Server) Stop() {
	self.Log.Notice("stopping Kubernetes server")
	close(self.stop)
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
