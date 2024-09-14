package client

import (
	"sync"
	"time"

	api "github.com/nephio-experimental/tko/api/grpc"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//
// Client
//

type Client struct {
	GRPCLevel2Protocol string
	GRPCAddress        string
	GRPCPort           int
	PackageFormat      string
	Timeout            time.Duration
	Timezone           *time.Location

	dataClient     api.DataClient
	dataClientLock sync.Mutex
	log            commonlog.Logger
}

func NewClient(grpcIpStack util.IPStack, grpcAddress string, grpcPort int, packageFormat string, timeout time.Duration, log commonlog.Logger) *Client {
	bind := grpcIpStack.ClientBind(grpcAddress)

	if address, err := util.ToReachableIPAddress(bind.Address); err == nil {
		bind.Address = address
	}

	return &Client{
		GRPCLevel2Protocol: bind.Level2Protocol,
		GRPCAddress:        bind.Address,
		GRPCPort:           grpcPort,
		PackageFormat:      packageFormat,
		Timeout:            timeout,
		Timezone:           time.Local,
		log:                log,
	}
}

func (self *Client) DataClient() (api.DataClient, error) {
	self.dataClientLock.Lock()
	defer self.dataClientLock.Unlock()

	if self.dataClient == nil {
		if clientConn, err := tkoutil.DialGRPCInsecure(self.GRPCAddress, self.GRPCPort); err == nil {
			self.dataClient = api.NewDataClient(clientConn)
		} else {
			return nil, err
		}
	}

	return self.dataClient, nil
}

// Utils

func (self *Client) encodePackage(package_ tkoutil.Package) ([]byte, error) {
	return tkoutil.EncodePackage(self.PackageFormat, package_)
}

func (self *Client) toTime(timestamp *timestamppb.Timestamp) time.Time {
	return timestamp.AsTime().In(self.Timezone)
}
