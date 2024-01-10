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

	grpcServers []*grpc.Server
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
	grpcServer := grpc.NewServer()
	api.RegisterAPIServer(grpcServer, self)

	protocol, address = tkoutil.GRPCDefaults(protocol, address)
	if address, addressZone, err := util.ToReachableIPAddress(address); err == nil {
		if addressZone != "" {
			address += "%" + addressZone
		}
		if listener, err := net.Listen(protocol, util.JoinIPAddressPort(address, self.Port)); err == nil {
			self.Log.Noticef("starting gRPC server %d on %s %s", len(self.grpcServers), protocol, listener.Addr().String())
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
