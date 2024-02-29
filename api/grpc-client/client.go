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

	apiClient     api.APIClient
	apiClientLock sync.Mutex
	log           commonlog.Logger
}

func NewClient(grpcIpStack util.IPStack, grpcAddress string, grpcPort int, resourcesFormat string, timeout time.Duration, log commonlog.Logger) *Client {
	bind := grpcIpStack.ClientBind(grpcAddress)

	if address, err := util.ToReachableIPAddress(bind.Address); err == nil {
		bind.Address = address
	}

	return &Client{
		GRPCLevel2Protocol: bind.Level2Protocol,
		GRPCAddress:        bind.Address,
		GRPCPort:           grpcPort,
		ResourcesFormat:    resourcesFormat,
		Timeout:            timeout,
		log:                log,
	}
}

func (self *Client) APIClient() (api.APIClient, error) {
	self.apiClientLock.Lock()
	defer self.apiClientLock.Unlock()

	if self.apiClient == nil {
		if clientConn, err := tkoutil.DialGRPCInsecure(self.GRPCAddress, self.GRPCPort); err == nil {
			self.apiClient = api.NewAPIClient(clientConn)
		} else {
			return nil, err
		}
	}

	return self.apiClient, nil
}

// Utils

func (self *Client) encodeResources(resources tkoutil.Resources) ([]byte, error) {
	return tkoutil.EncodeResources(self.ResourcesFormat, resources)
}
