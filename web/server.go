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

	httpServer *http.Server
	mux        *http.ServeMux
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

	self.httpServer = &http.Server{
		Handler: self.mux,
	}

	return &self, nil
}

func (self *Server) Start() error {
	if listener, err := net.Listen(self.Protocol, util.JoinIPAddressPort(self.Address, self.Port)); err == nil {
		self.Log.Noticef("starting web server on %s", listener.Addr().String())
		go func() {
			if err := self.httpServer.Serve(listener); err != nil {
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

func (self *Server) Stop() error {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), 5*time.Second)
	defer cancel()

	self.Log.Notice("stopping web server")
	return self.httpServer.Shutdown(context)
}
