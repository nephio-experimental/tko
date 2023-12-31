package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
)

type PluginInput struct {
	GRPC                    PluginInputGRPC         `yaml:"grpc"`
	LogFile                 string                  `yaml:"logFile"`
	Resources               util.Resources          `yaml:"resources"`
	TargetResourceIdentifer util.ResourceIdentifier `yaml:"targetResourceIdentifier"`
	Complete                bool                    `yaml:"complete"`
}

type PluginInputGRPC struct {
	Protocol string `yaml:"protocol"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
}

type PluginOutput struct {
	Error string `yaml:"error,omitempty"`
}

func (self *Context) ToPluginInput(logFile string) PluginInput {
	return PluginInput{
		GRPC: PluginInputGRPC{
			Protocol: self.Validation.Client.GRPCProtocol,
			Address:  self.Validation.Client.GRPCAddress,
			Port:     self.Validation.Client.GRPCPort,
		},
		LogFile:                 logFile,
		Resources:               self.Resources,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
		Complete:                self.Complete,
	}
}

func NewPluginValidator(plugin client.PluginInfo) (ValidatorFunc, error) {
	switch plugin.Executor {
	case "command":
		return NewCommandPluginValidator(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", plugin.Type)
	}
}

func NewCommandPluginValidator(plugin client.PluginInfo) (ValidatorFunc, error) {
	if len(plugin.Arguments) < 1 {
		return nil, errors.New("plugin of type \"command\" must have at least one argument")
	}

	return func(context *Context) error {
		context.Validation.Log.Infof("validate via command plugin for %s: %s", context.TargetResourceIdentifer, strings.Join(plugin.Arguments, " "))

		logFifo := util.NewLogFIFO("tko-validation", context.Validation.Log)
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
