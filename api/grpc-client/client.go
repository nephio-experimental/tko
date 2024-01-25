package client

import (
	"sync"
	"time"

	api "github.com/nephio-experimental/tko/api/grpc"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

//
// Client
//

type Client struct {
	GRPCLevel2Protocol string
	GRPCAddress        string
	GRPCPort           int
	ResourcesFormat    string
	Timeout            time.Duration

	apiClient_    api.APIClient
	apiClientLock sync.Mutex
	log           commonlog.Logger
}

func NewClient(grpcIpStack util.IPStack, grpcAddress string, grpcPort int, resourcesFormat string, timeoutSeconds float64, log commonlog.Logger) (*Client, error) {
	bind := grpcIpStack.ClientBind(grpcAddress)

	if address, err := util.ToReachableIPAddress(bind.Address); err == nil {
		return &Client{
			GRPCLevel2Protocol: bind.Level2Protocol,
			GRPCAddress:        address,
			GRPCPort:           grpcPort,
			ResourcesFormat:    resourcesFormat,
			Timeout:            time.Duration(timeoutSeconds * float64(time.Second)),
			log:                log,
		}, nil
	} else {
		return nil, err
	}
}

func (self *Client) apiClient() (api.APIClient, error) {
	self.apiClientLock.Lock()
	defer self.apiClientLock.Unlock()

	if self.apiClient_ == nil {
		if clientConn, err := tkoutil.DialGRPCInsecure(self.GRPCAddress, self.GRPCPort); err == nil {
			self.apiClient_ = api.NewAPIClient(clientConn)
		} else {
			return nil, err
		}
	}

	return self.apiClient_, nil
}

// Utils

func (self *Client) encodeResources(resources tkoutil.Resources) ([]byte, error) {
	return tkoutil.EncodeResources(self.ResourcesFormat, resources)
}
