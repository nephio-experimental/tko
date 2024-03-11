package server

import (
	contextpkg "context"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/tliron/kutil/version"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ([api.APIServer] interface)
func (self *Server) About(context contextpkg.Context, _ *emptypb.Empty) (*api.AboutResponse, error) {
	self.Log.Info("about")

	return &api.AboutResponse{
		InstanceName:         self.InstanceName,
		InstanceDescription:  self.InstanceDescription,
		TkoVersion:           version.GitVersion,
		Backend:              self.Backend.String(),
		IpStack:              string(self.IPStack),
		AddressPorts:         self.clientAddressPorts,
		DefaultPackageFormat: self.DefaultPackageFormat,
	}, nil
}
