package server

import (
	contextpkg "context"
	"net"
	"net/http"
	"time"

	"github.com/nephio-experimental/tko/assets/web"
	"github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

//
// Server
//

type Server struct {
	InstanceName        string
	InstanceDescription string
	Backend             backend.Backend
	Timeout             time.Duration
	IPStack             util.IPStack
	Address             string
	Port                int
	Timezone            *time.Location
	Log                 commonlog.Logger

	httpServers []*http.Server
	addresses   []string
	mux         *http.ServeMux
}

func NewServer(backend backend.Backend, timeout time.Duration, ipStack util.IPStack, address string, port int, timezone *time.Location, log commonlog.Logger) (*Server, error) {
	if timezone == nil {
		timezone = time.Local
	}

	self := Server{
		Backend:  backend,
		Timeout:  timeout,
		IPStack:  ipStack,
		Address:  address,
		Port:     port,
		Timezone: timezone,
		Log:      log,
		mux:      http.NewServeMux(),
	}

	self.mux.Handle("/", http.FileServer(http.FS(web.FS)))

	self.mux.HandleFunc("/api/about", self.About)
	self.mux.HandleFunc("/api/deployment/list", self.ListDeployments)
	self.mux.HandleFunc("/api/deployment", self.GetDeployment)
	self.mux.HandleFunc("/api/site/list", self.ListSites)
	self.mux.HandleFunc("/api/site", self.GetSite)
	self.mux.HandleFunc("/api/template/list", self.ListTemplates)
	self.mux.HandleFunc("/api/template", self.GetTemplate)
	self.mux.HandleFunc("/api/plugin/list", self.ListPlugins)

	return &self, nil
}

func (self *Server) Start() error {
	return tkoutil.StartServer(self.IPStack, self.Address, self.start)
}

// ([util.StartServerFunc] signature)
func (self *Server) start(level2protocol string, address string) error {
	address = util.JoinIPAddressPort(address, self.Port)
	if listener, err := net.Listen(level2protocol, address); err == nil {
		index := len(self.httpServers)
		self.Log.Notice("starting HTTP server",
			"index", index,
			"level2protocol", level2protocol,
			"address", listener.Addr().String())

		httpServer := http.Server{
			Handler: http.TimeoutHandler(self.mux, self.Timeout, ""),
		}
		self.httpServers = append(self.httpServers, &httpServer)
		self.addresses = append(self.addresses, address)

		go func() {
			if err := httpServer.Serve(listener); err != nil {
				if err == http.ErrServerClosed {
					self.Log.Info("stopped HTTP server",
						"index", index)
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
		self.Log.Notice("stopping HTTP server",
			"index", index)
		if err := httpServer.Shutdown(context); err != nil {
			self.Log.Critical(err.Error())
		}
	}
}
