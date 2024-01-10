package web

import (
	contextpkg "context"
	"net"
	"net/http"
	"time"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/assets/web"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

//
// Server
//

type Server struct {
	Backend  backend.Backend
	Protocol string
	Address  string
	Port     int
	Log      commonlog.Logger

	httpServers []*http.Server
	mux         *http.ServeMux
}

func NewServer(backend backend.Backend, protocol string, address string, port int, log commonlog.Logger) (*Server, error) {
	self := Server{
		Backend:  backend,
		Protocol: protocol,
		Address:  address,
		Port:     port,
		Log:      log,
		mux:      http.NewServeMux(),
	}

	self.mux.Handle("/", http.FileServer(http.FS(web.FS)))

	self.mux.HandleFunc("/api/deployment/list", self.listDeployments)
	self.mux.HandleFunc("/api/deployment", self.getDeployment)
	self.mux.HandleFunc("/api/site/list", self.listSites)
	self.mux.HandleFunc("/api/site", self.getSite)
	self.mux.HandleFunc("/api/template/list", self.listTemplates)
	self.mux.HandleFunc("/api/template", self.getTemplate)
	self.mux.HandleFunc("/api/plugin/list", self.listPlugins)

	return &self, nil
}

func (self *Server) Start() error {
	if ((self.Protocol == "tcp") || (self.Protocol == "")) && (self.Address == "") {
		// For dual stack "bind all" (empty address) we need to bind separately for each protocol
		// See: https://github.com/golang/go/issues/9334
		if err := self.start("tcp6", ""); err != nil {
			return err
		}
		return self.start("tcp4", "")
	} else {
		return self.start(self.Protocol, self.Address)
	}
}

func (self *Server) start(protocol string, address string) error {
	httpServer := http.Server{
		Handler: self.mux,
	}

	if listener, err := net.Listen(protocol, util.JoinIPAddressPort(address, self.Port)); err == nil {
		self.Log.Noticef("starting web server %d on %s %s", len(self.httpServers), protocol, listener.Addr().String())
		self.httpServers = append(self.httpServers, &httpServer)
		go func() {
			if err := httpServer.Serve(listener); err != nil {
				if err == http.ErrServerClosed {
					self.Log.Info("stopped web server")
				} else {
					self.Log.Errorf("%s", err.Error())
				}
			}
		}()
		return nil
	} else {
		return err
	}
}

func (self *Server) Stop() {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), 5*time.Second)
	defer cancel()

	for index, httpServer := range self.httpServers {
		self.Log.Noticef("stopping web server %d", index)
		if err := httpServer.Shutdown(context); err != nil {
			self.Log.Critical(err.Error())
		}
		self.Log.Noticef("stopped web server %d", index)
	}
}
