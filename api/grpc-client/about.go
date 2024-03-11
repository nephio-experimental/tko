package client

import (
	contextpkg "context"

	"google.golang.org/protobuf/types/known/emptypb"
)

type About struct {
	InstanceName        string    `json:"instanceName" yaml:"instanceName"`
	InstanceDescription string    `json:"instanceDescription" yaml:"instanceDescription"`
	TKOVersion          string    `json:"tkoVersion" yaml:"tkoVersion"`
	Backend             string    `json:"backend" yaml:"backend"`
	GRPC                AboutGRPC `json:"grpc" yaml:"grpc"`
}

type AboutGRPC struct {
	IPStack              string   `json:"ipStack" yaml:"ipStack"`
	AddressPorts         []string `json:"addressPorts" yaml:"addressPorts"`
	DefaultPackageFormat string   `json:"defaultPackageFormat" yaml:"defaultPackageFormat"`
}

func (self *Client) About() (About, error) {
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Infof("about")
		if response, err := apiClient.About(context, new(emptypb.Empty)); err == nil {
			return About{
				InstanceName:        response.InstanceName,
				InstanceDescription: response.InstanceDescription,
				TKOVersion:          response.TkoVersion,
				Backend:             response.Backend,
				GRPC: AboutGRPC{
					IPStack:              response.IpStack,
					AddressPorts:         response.AddressPorts,
					DefaultPackageFormat: response.DefaultPackageFormat,
				},
			}, nil
		} else {
			return About{}, err
		}
	} else {
		return About{}, err
	}
}
