package server

import (
	"net"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc"
)

//
// GRPCServer
//

type Server struct {
	api.UnimplementedAPIServer

	InstanceName         string
	InstanceDescription  string
	Backend              backend.Backend
	IPStack              util.IPStack
	Address              string
	Port                 int
	DefaultPackageFormat string
	Log                  commonlog.Logger

	grpcServers     []*grpc.Server
	clientAddresses []string
}

func NewServer(backend backend.Backend, ipStack util.IPStack, address string, port int, defaultPackageFormat string, log commonlog.Logger) *Server {
	return &Server{
		Backend:              backend,
		IPStack:              ipStack,
		Address:              address,
		Port:                 port,
		DefaultPackageFormat: defaultPackageFormat,
		Log:                  log,
	}
}

func (self *Server) Start() error {
	return self.IPStack.StartServers(self.Address, self.start)
}

func (self *Server) Stop() {
	for index, grpcServer := range self.grpcServers {
		self.Log.Notice("stopping gRPC server",
			"index", index)
		grpcServer.Stop() // TODO: GracefulStop()?
		self.Log.Notice("stopped gRPC server",
			"index", index)
	}
}

// ([util.IPStackStartServerFunc] signature)
func (self *Server) start(level2protocol string, address string) error {
	if address, err := util.ToReachableIPAddress(address); err == nil {
		addressPort := util.JoinIPAddressPort(address, self.Port)
		if listener, err := net.Listen(level2protocol, addressPort); err == nil {
			self.Log.Notice("starting gRPC server",
				"index", len(self.grpcServers),
				"level2protocol", level2protocol,
				"addressPort", listener.Addr().String())

			grpcServer := grpc.NewServer()
			api.RegisterAPIServer(grpcServer, self)
			self.grpcServers = append(self.grpcServers, grpcServer)
			self.clientAddresses = append(self.clientAddresses, util.IPAddressPortWithoutZone(addressPort))

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
