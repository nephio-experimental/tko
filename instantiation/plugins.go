package instantiation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
)

type PluginInput struct {
	GRPC                    PluginInputGRPC           `yaml:"grpc"`
	LogFile                 string                    `yaml:"logFile"`
	SiteID                  string                    `yaml:"siteId"`
	SiteResources           util.Resources            `yaml:"siteResources"`
	TargetResourceIdentifer util.ResourceIdentifier   `yaml:"targetResourceIdentifier"`
	Deployments             map[string]util.Resources `yaml:"deployments"`
}

type PluginInputGRPC struct {
	Level2Protocol string `yaml:"level2protocol"`
	Address        string `yaml:"address"`
	Port           int    `yaml:"port"`
}

type PluginOutput struct {
	Error string `yaml:"error,omitempty"`
}

func (self *Context) ToPluginInput(logFile string) PluginInput {
	return PluginInput{
		GRPC: PluginInputGRPC{
			Level2Protocol: self.Instantiation.Client.GRPCLevel2Protocol,
			Address:        self.Instantiation.Client.GRPCAddress,
			Port:           self.Instantiation.Client.GRPCPort,
		},
		LogFile:                 logFile,
		SiteID:                  self.SiteID,
		SiteResources:           self.SiteResources,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
		Deployments:             self.Deployments,
	}
}

func NewPluginInstantiator(plugin client.PluginInfo) (InstantiatorFunc, error) {
	switch plugin.Executor {
	case "command":
		return NewCommandPluginInstantiator(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", plugin.Type)
	}
}

func NewCommandPluginInstantiator(plugin client.PluginInfo) (InstantiatorFunc, error) {
	if len(plugin.Arguments) < 1 {
		return nil, errors.New("plugin of type \"command\" must have at least one argument")
	}

	return func(context *Context) error {
		context.Log.Infof("instantiate via command plugin for %s: %s", context.TargetResourceIdentifer, strings.Join(plugin.Arguments, " "))

		logFifo := util.NewLogFIFO("tko-instantiation", context.Log)
		if err := logFifo.Start(); err != nil {
			return err
		}

		input := context.ToPluginInput(logFifo.Path)
		var output PluginOutput
		if err := util.ExecuteCommand(plugin.Arguments, input, &output); err == nil {
			if output.Error == "" {
				return nil
			} else {
				return errors.New(output.Error)
			}
		} else {
			return err
		}
	}, nil
}
