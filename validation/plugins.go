package validation

import (
	"errors"
	"fmt"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
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
			Level2Protocol: self.Validation.Client.GRPCLevel2Protocol,
			Address:        self.Validation.Client.GRPCAddress,
			Port:           self.Validation.Client.GRPCPort,
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

	return func(validationContext *Context) []error {
		validationContext.Validation.Log.Info("validate via command plugin",
			"resource", validationContext.TargetResourceIdentifer,
			"arguments", strings.Join(plugin.Arguments, " "))

		logFifo := util.NewLogFIFO("tko-validation", validationContext.Validation.Log)
		if err := logFifo.Start(); err != nil {
			return []error{err}
		}

		input := validationContext.ToPluginInput(logFifo.Path)
		var output PluginOutput
		if err := util.ExecuteCommand(plugin.Arguments, input, &output); err == nil {
			if output.Error == "" {
				return nil
			} else {
				return []error{errors.New(output.Error)}
			}
		} else {
			return []error{err}
		}
	}, nil
}
