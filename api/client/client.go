package client

import (
	api "github.com/nephio-experimental/tko/grpc"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//
// Client
//

type Client struct {
	GRPCProtocol    string
	GRPCAddress     string
	GRPCPort        int
	ResourcesFormat string

	client api.APIClient
	log    commonlog.Logger
}

func NewClient(grpcProtocol string, grpcAddress string, grpcPort int, resourcesFormat string, log commonlog.Logger) (*Client, error) {
	_, grpcAddress = tkoutil.GRPCDefaults(grpcProtocol, grpcAddress)
	if grpcAddress, grpcAddressZone, err := util.ToReachableIPAddress(grpcAddress); err == nil {
		if grpcAddressZone != "" {
			// See: https://github.com/grpc/grpc-go/issues/3272#issuecomment-1239710027
			grpcAddress += "%25" + grpcAddressZone
		}
		if clientConn, err := grpc.Dial(util.JoinIPAddressPort(grpcAddress, grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials())); err == nil {
			return &Client{
				GRPCProtocol:    grpcProtocol,
				GRPCAddress:     grpcAddress,
				GRPCPort:        grpcPort,
				ResourcesFormat: resourcesFormat,
				client:          api.NewAPIClient(clientConn),
				log:             log,
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// Utils

func (self *Client) encodeResources(resources []tkoutil.Resource) ([]byte, error) {
	return tkoutil.EncodeResources(self.ResourcesFormat, resources)
}
