package server

import (
	"net"

	"github.com/nephio-experimental/tko/api/backend"
	api "github.com/nephio-experimental/tko/grpc"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc"
)

//
// GRPCServer
//

type Server struct {
	api.UnimplementedAPIServer

	Backend                backend.Backend
	IPStack                string
	Address                string
	Port                   int
	DefaultResourcesFormat string
	Log                    commonlog.Logger

	grpcServers []*grpc.Server
}

func NewServer(backend backend.Backend, ipStack string, address string, port int, defaultResourcesFormat string, log commonlog.Logger) *Server {
	return &Server{
		Backend:                backend,
		IPStack:                ipStack,
		Address:                address,
		Port:                   port,
		DefaultResourcesFormat: defaultResourcesFormat,
		Log:                    log,
	}
}

func (self *Server) Start() error {
	return tkoutil.StartServer(self.IPStack, self.Address, self.start)
}

// ([util.StartServerFunc] signature)
func (self *Server) start(level2protocol string, address string) error {
	grpcServer := grpc.NewServer()
	api.RegisterAPIServer(grpcServer, self)

	if address, err := tkoutil.ToReachableIPAddress(address); err == nil {
		if listener, err := net.Listen(level2protocol, util.JoinIPAddressPort(address, self.Port)); err == nil {
			self.Log.Noticef("starting gRPC server %d on %s %s", len(self.grpcServers), level2protocol, listener.Addr().String())
			self.grpcServers = append(self.grpcServers, grpcServer)
			go func() {
				if err := grpcServer.Serve(listener); err != nil {
					self.Log.Critical(err.Error())
				}
			}()
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *Server) Stop() {
	for index, grpcServer := range self.grpcServers {
		if grpcServer != nil {
			self.Log.Noticef("stopping gRPC server %d", index)
			grpcServer.Stop()
			self.Log.Noticef("stopped gRPC server %d", index)
		}
	}
}
