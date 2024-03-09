package preparation

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	pluginspkg "github.com/nephio-experimental/tko/plugins"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

const FIFOPrefix = "tko-preparation-"

type PluginInput struct {
	GRPC                    PluginInputGRPC            `yaml:"grpc"`
	LogFile                 string                     `yaml:"logFile"`
	LogAddressPort          string                     `yaml:"logAddressPort"`
	DeploymentID            string                     `yaml:"deploymentId"`
	DeploymentPackage       tkoutil.Package            `yaml:"deploymentPackage"`
	TargetResourceIdentifer tkoutil.ResourceIdentifier `yaml:"targetResourceIdentifier"`
}

type PluginInputGRPC struct {
	Level2Protocol string `yaml:"level2protocol"`
	Address        string `yaml:"address"`
	Port           int    `yaml:"port"`
}

type PluginOutput struct {
	Prepared bool            `yaml:"prepared,omitempty"`
	Package  tkoutil.Package `yaml:"package,omitempty"`
	Error    string          `yaml:"error,omitempty"`
}

func (self *Context) ToPluginInput(logFile string, logAddressPort string) PluginInput {
	return PluginInput{
		GRPC: PluginInputGRPC{
			Level2Protocol: self.Preparation.Client.GRPCLevel2Protocol,
			Address:        self.Preparation.Client.GRPCAddress,
			Port:           self.Preparation.Client.GRPCPort,
		},
		LogFile:                 logFile,
		LogAddressPort:          logAddressPort,
		DeploymentID:            self.DeploymentID,
		DeploymentPackage:       self.DeploymentPackage,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
	}
}

func NewPluginPreparer(plugin client.Plugin, logIpStack util.IPStack, logAddress string, logPort int) (PrepareFunc, error) {
	switch plugin.Executor {
	case pluginspkg.Command:
		return NewCommandPluginPreparer(plugin, logIpStack, logAddress, logPort)
	case pluginspkg.Kpt:
		return NewKptPluginPreparer(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin executor: %s", plugin.Executor)
	}
}

func NewCommandPluginPreparer(plugin client.Plugin, logIpStack util.IPStack, logAddress string, logPort int) (PrepareFunc, error) {
	executor, err := pluginspkg.NewCommandExecutor(plugin.Arguments, plugin.Properties)
	if err != nil {
		return nil, err
	}

	return func(context contextpkg.Context, preparationContext *Context) (bool, []ard.Map, error) {
		preparationContext.Log.Info("prepare via command plugin",
			"resource", preparationContext.TargetResourceIdentifer,
			"arguments", strings.Join(plugin.Arguments, " "))

		var input PluginInput
		var output PluginOutput

		if logFile, logAddressPort, err := executor.GetLog(FIFOPrefix, logIpStack, logAddress, logPort, preparationContext.Preparation.Log); err == nil {
			input = preparationContext.ToPluginInput(logFile, logAddressPort)
		} else {
			return false, nil, err
		}

		if err := executor.Execute(context, input, &output); err == nil {
			if output.Error == "" {
				return output.Prepared, output.Package, nil
			} else {
				return false, nil, errors.New(output.Error)
			}
		} else {
			return false, nil, err
		}
	}, nil
}

func NewKptPluginPreparer(plugin client.Plugin) (PrepareFunc, error) {
	executor, err := pluginspkg.NewKptExecutor(plugin.Arguments, plugin.Properties)
	if err != nil {
		return nil, err
	}

	return func(context contextpkg.Context, preparationContext *Context) (bool, []ard.Map, error) {
		if package_, err := executor.Execute(context, preparationContext.TargetResourceIdentifer, preparationContext.DeploymentPackage); err == nil {
			// Note: it's OK if the kpt function deleted our plugin resource because that also counts as completion
			if resource, ok := preparationContext.TargetResourceIdentifer.GetResource(package_); ok {
				if !tkoutil.SetPreparedAnnotation(resource, true) {
					return false, nil, errors.New("malformed resource")
				}
			}
			return true, package_, nil
		} else {
			return false, nil, err
		}
	}, nil
}
