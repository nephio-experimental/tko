package server

import (
	"net"

	"github.com/nephio-experimental/tko/api/backend"
	api "github.com/nephio-experimental/tko/api/grpc"
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
	IPStack                util.IPStack
	Address                string
	Port                   int
	DefaultResourcesFormat string
	Log                    commonlog.Logger

	grpcServers []*grpc.Server
}

func NewServer(backend backend.Backend, ipStack util.IPStack, address string, port int, defaultResourcesFormat string, log commonlog.Logger) *Server {
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
	if address, err := util.ToReachableIPAddress(address); err == nil {
		address = util.JoinIPAddressPort(address, self.Port)
		if listener, err := net.Listen(level2protocol, address); err == nil {
			self.Log.Notice("starting gRPC server",
				"index", len(self.grpcServers),
				"level2protocol", level2protocol,
				"address", listener.Addr().String())

			grpcServer := grpc.NewServer()
			api.RegisterAPIServer(grpcServer, self)
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
			self.Log.Notice("stopping gRPC server",
				"index", index)
			grpcServer.Stop() // TODO: GracefulStop()?
			self.Log.Notice("stopped gRPC server",
				"index", index)
		}
	}
}
