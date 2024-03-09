package validation

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	pluginspkg "github.com/nephio-experimental/tko/plugins"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

const FIFOPrefix = "tko-validation-"

type PluginInput struct {
	GRPC                    PluginInputGRPC            `yaml:"grpc"`
	LogFile                 string                     `yaml:"logFile"`
	LogAddressPort          string                     `yaml:"logAddressPort"`
	Package                 tkoutil.Package            `yaml:"package"`
	TargetResourceIdentifer tkoutil.ResourceIdentifier `yaml:"targetResourceIdentifier"`
	Complete                bool                       `yaml:"complete"`
}

type PluginInputGRPC struct {
	Level2Protocol string `yaml:"level2protocol"`
	Address        string `yaml:"address"`
	Port           int    `yaml:"port"`
}

type PluginOutput struct {
	Error string `yaml:"error,omitempty"`
}

func (self *Context) ToPluginInput(logFile string, logAddressPort string) PluginInput {
	return PluginInput{
		GRPC: PluginInputGRPC{
			Level2Protocol: self.Validation.Client.GRPCLevel2Protocol,
			Address:        self.Validation.Client.GRPCAddress,
			Port:           self.Validation.Client.GRPCPort,
		},
		LogFile:                 logFile,
		LogAddressPort:          logAddressPort,
		Package:                 self.Package,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
		Complete:                self.Complete,
	}
}

func NewPluginValidator(plugin client.Plugin, logIpStack util.IPStack, logAddress string, logPort int) (ValidateFunc, error) {
	switch plugin.Executor {
	case pluginspkg.Command:
		return NewCommandPluginValidator(plugin, logIpStack, logAddress, logPort)
	default:
		return nil, fmt.Errorf("unsupported plugin executor: %s", plugin.Executor)
	}
}

func NewCommandPluginValidator(plugin client.Plugin, logIpStack util.IPStack, logAddress string, logPort int) (ValidateFunc, error) {
	executor, err := pluginspkg.NewCommandExecutor(plugin.Arguments, plugin.Properties)
	if err != nil {
		return nil, err
	}

	return func(context contextpkg.Context, validationContext *Context) []error {
		validationContext.Validation.Log.Info("validate via command plugin",
			"resource", validationContext.TargetResourceIdentifer,
			"arguments", strings.Join(plugin.Arguments, " "))

		var input PluginInput
		var output PluginOutput

		if logFile, logAddressPort, err := executor.GetLog(FIFOPrefix, logIpStack, logAddress, logPort, validationContext.Validation.Log); err == nil {
			input = validationContext.ToPluginInput(logFile, logAddressPort)
		} else {
			return []error{err}
		}

		if err := executor.Execute(context, input, &output); err == nil {
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
