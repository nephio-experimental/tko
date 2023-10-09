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
	Protocol               string
	Address                string
	Port                   int
	DefaultResourcesFormat string
	Log                    commonlog.Logger

	grpcServer *grpc.Server
}

func NewServer(backend backend.Backend, protocol string, address string, port int, defaultResourcesFormat string, log commonlog.Logger) *Server {
	return &Server{
		Backend:                backend,
		Protocol:               protocol,
		Address:                address,
		Port:                   port,
		DefaultResourcesFormat: defaultResourcesFormat,
		Log:                    log,
	}
}

func (self *Server) Start() error {
	self.grpcServer = grpc.NewServer()
	api.RegisterAPIServer(self.grpcServer, self)

	protocol, address := tkoutil.GRPCDefaults(self.Protocol, self.Address)
	if address, addressZone, err := util.ToReachableIPAddress(address); err == nil {
		if addressZone != "" {
			address += "%" + addressZone
		}
		if listener, err := net.Listen(protocol, util.JoinIPAddressPort(address, self.Port)); err == nil {
			self.Log.Noticef("starting gRPC server on %s", listener.Addr().String())
			go func() {
				if err := self.grpcServer.Serve(listener); err != nil {
					self.Log.Criticalf("%s", err.Error())
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
	if self.grpcServer != nil {
		self.Log.Notice("stopping gRPC server")
		self.grpcServer.Stop()
		self.Log.Notice("stopped gRPC server")
	}
}
