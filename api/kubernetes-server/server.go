package server

import (
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	"k8s.io/apiserver/pkg/server"
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
}

func NewServer(backend backend.Backend, port int, log commonlog.Logger) (*Server, error) {
	server := Server{
		Backend: backend,
		Log:     log,
		stop:    make(chan struct{}),
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
	go func() {
		if err := self.apiServer.PrepareRun().Run(self.stop); err != nil {
			self.Log.Error(err.Error())
		}
	}()
}

func (self *Server) Stop() {
	self.stop <- struct{}{}
}
