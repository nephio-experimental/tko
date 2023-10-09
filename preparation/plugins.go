package preparation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

type PluginInput struct {
	GRPC                    PluginInputGRPC         `yaml:"grpc"`
	LogFile                 string                  `yaml:"logFile"`
	DeploymentID            string                  `yaml:"deploymentId"`
	DeploymentResources     []util.Resource         `yaml:"deploymentResources"`
	TargetResourceIdentifer util.ResourceIdentifier `yaml:"targetResourceIdentifier"`
}

type PluginInputGRPC struct {
	Protocol string `yaml:"protocol"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
}

type PluginOutput struct {
	Prepared  bool            `yaml:"prepared,omitempty"`
	Resources []util.Resource `yaml:"resources,omitempty"`
	Error     string          `yaml:"error,omitempty"`
}

func (self *Context) ToPluginInput(logFile string) PluginInput {
	return PluginInput{
		GRPC: PluginInputGRPC{
			Protocol: self.Preparation.Client.GRPCProtocol,
			Address:  self.Preparation.Client.GRPCAddress,
			Port:     self.Preparation.Client.GRPCPort,
		},
		LogFile:                 logFile,
		DeploymentID:            self.DeploymentID,
		DeploymentResources:     self.DeploymentResources,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
	}
}

func NewPluginPreparer(plugin client.PluginInfo) (PreparerFunc, error) {
	switch plugin.Executor {
	case "command":
		return NewCommandPluginPreparer(plugin)
	case "kpt":
		return NewKptPluginPreparer(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", plugin.Type)
	}
}

func NewCommandPluginPreparer(plugin client.PluginInfo) (PreparerFunc, error) {
	if len(plugin.Arguments) < 1 {
		return nil, errors.New("plugin of type \"command\" must have at least one argument")
	}

	return func(context *Context) (bool, []ard.Map, error) {
		context.Log.Infof("prepare via command plugin for %s: %s", context.TargetResourceIdentifer, strings.Join(plugin.Arguments, " "))

		logFifo := util.NewLogFIFO("tko-preparation", context.Log)
		if err := logFifo.Start(); err != nil {
			return false, nil, err
		}

		input := context.ToPluginInput(logFifo.Path)
		var output PluginOutput
		if err := util.ExecuteCommand(plugin.Arguments, input, &output); err == nil {
			if output.Error == "" {
				return output.Prepared, output.Resources, nil
			} else {
				return false, nil, errors.New(output.Error)
			}
		} else {
			return false, nil, err
		}
	}, nil
}

func NewKptPluginPreparer(plugin client.PluginInfo) (PreparerFunc, error) {
	if len(plugin.Arguments) != 1 {
		return nil, errors.New("plugin of type \"command\" must have one argument")
	}

	image := plugin.Arguments[0]

	return func(context *Context) (bool, []ard.Map, error) {
		context.Log.Infof("prepare via kpt plugin for %s: %s, %s", context.TargetResourceIdentifer, image, plugin.Properties)

		if resource, ok := context.GetResource(); ok {
			//context.Log.Noticef("!!! %s", resource)
			if resources, err := util.ExecuteKpt(image, plugin.Properties, resource, context.DeploymentResources); err == nil {
				// Note: it's OK if the kpt function deleted our plugin resource because that also counts as completion
				if resource_, ok := context.TargetResourceIdentifer.GetResource(resources); ok {
					if !util.SetPreparedAnnotation(resource_, true) {
						return false, nil, errors.New("malformed resource")
					}
				}

				return true, resources, nil
			} else {
				return false, nil, err
			}
		} else {
			// Our resource disappeared
			return false, nil, nil
		}
	}, nil
}
