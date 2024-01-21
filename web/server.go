package web

import (
	contextpkg "context"
	"net"
	"net/http"
	"time"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/assets/web"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

//
// Server
//

type Server struct {
	Backend backend.Backend
	IPStack util.IPStack
	Address string
	Port    int
	Log     commonlog.Logger

	httpServers []*http.Server
	mux         *http.ServeMux
}

func NewServer(backend backend.Backend, ipStack util.IPStack, address string, port int, log commonlog.Logger) (*Server, error) {
	self := Server{
		Backend: backend,
		IPStack: ipStack,
		Address: address,
		Port:    port,
		Log:     log,
		mux:     http.NewServeMux(),
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
	return tkoutil.StartServer(self.IPStack, self.Address, self.start)
}

// ([util.StartServerFunc] signature)
func (self *Server) start(level2protocol string, address string) error {
	address = util.JoinIPAddressPort(address, self.Port)
	if listener, err := net.Listen(level2protocol, address); err == nil {
		self.Log.Noticef("starting web server %d on %s %s", len(self.httpServers), level2protocol, listener.Addr().String())

		httpServer := http.Server{
			Handler: self.mux,
		}
		self.httpServers = append(self.httpServers, &httpServer)

		go func() {
			if err := httpServer.Serve(listener); err != nil {
				if err == http.ErrServerClosed {
					self.Log.Info("stopped web server")
				} else {
					self.Log.Error(err.Error())
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
