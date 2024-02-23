package server

import (
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	"k8s.io/apiserver/pkg/server"

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
}

func NewServer(backend backend.Backend, port int, log commonlog.Logger) (*Server, error) {
	config, err := NewConfig(port)
	if err != nil {
		return nil, err
	}

	apiGroupInfo, err := NewAPIGroupInfo(config.RESTOptionsGetter, backend, log)
	if err != nil {
		return nil, err
	}

	apiServer, err := config.New(ServerName, server.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	if err := apiServer.InstallAPIGroup(apiGroupInfo); err != nil {
		return nil, err
	}

	return &Server{
		Backend:   backend,
		Log:       log,
		apiServer: apiServer,
		stop:      make(chan struct{}),
	}, nil
}

func (self *Server) Start() {
	go func() {
		if err := self.apiServer.PrepareRun().Run(self.stop); err != nil {
			self.Log.Error(err.Error())
		}
	}()
}

func (self *Server) Stop() {
	self.stop <- struct{}{}
}
