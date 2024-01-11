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
	GRPCLevel2Protocol string
	GRPCAddress        string
	GRPCPort           int
	ResourcesFormat    string

	client api.APIClient
	log    commonlog.Logger
}

func NewClient(grpcIpStack string, grpcAddress string, grpcPort int, resourcesFormat string, log commonlog.Logger) (*Client, error) {
	var level2protocol string
	var err error
	if level2protocol, grpcAddress, err = tkoutil.IPLevel2ProtocolAndAddress(grpcIpStack, grpcAddress); err != nil {
		return nil, err
	}

	if grpcAddress, err := tkoutil.ToReachableIPAddress(grpcAddress); err == nil {
		if clientConn, err := grpc.Dial(util.JoinIPAddressPort(grpcAddress, grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials())); err == nil {
			return &Client{
				GRPCLevel2Protocol: level2protocol,
				GRPCAddress:        grpcAddress,
				GRPCPort:           grpcPort,
				ResourcesFormat:    resourcesFormat,
				client:             api.NewAPIClient(clientConn),
				log:                log,
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// Utils

func (self *Client) encodeResources(resources tkoutil.Resources) ([]byte, error) {
	return tkoutil.EncodeResources(self.ResourcesFormat, resources)
}
