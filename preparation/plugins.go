package preparation

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	pluginspkg "github.com/nephio-experimental/tko/plugins"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

const FIFOPrefix = "tko-preparation-"

type PluginInput struct {
	GRPC                    PluginInputGRPC         `yaml:"grpc"`
	LogFile                 string                  `yaml:"logFile"`
	DeploymentID            string                  `yaml:"deploymentId"`
	DeploymentPackage       util.Package            `yaml:"deploymentPackage"`
	TargetResourceIdentifer util.ResourceIdentifier `yaml:"targetResourceIdentifier"`
}

type PluginInputGRPC struct {
	Level2Protocol string `yaml:"level2protocol"`
	Address        string `yaml:"address"`
	Port           int    `yaml:"port"`
}

type PluginOutput struct {
	Prepared bool         `yaml:"prepared,omitempty"`
	Package  util.Package `yaml:"package,omitempty"`
	Error    string       `yaml:"error,omitempty"`
}

func (self *Context) ToPluginInput(logFile string) PluginInput {
	return PluginInput{
		GRPC: PluginInputGRPC{
			Level2Protocol: self.Preparation.Client.GRPCLevel2Protocol,
			Address:        self.Preparation.Client.GRPCAddress,
			Port:           self.Preparation.Client.GRPCPort,
		},
		LogFile:                 logFile,
		DeploymentID:            self.DeploymentID,
		DeploymentPackage:       self.DeploymentPackage,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
	}
}

func NewPluginPreparer(plugin client.Plugin) (PrepareFunc, error) {
	switch plugin.Executor {
	case pluginspkg.Command:
		return NewCommandPluginPreparer(plugin)
	case pluginspkg.Kpt:
		return NewKptPluginPreparer(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin executor: %s", plugin.Executor)
	}
}

func NewCommandPluginPreparer(plugin client.Plugin) (PrepareFunc, error) {
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

		if logFifo, err := executor.GetLogFIFO(FIFOPrefix, preparationContext.Log); err == nil {
			input = preparationContext.ToPluginInput(logFifo)
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
				if !util.SetPreparedAnnotation(resource, true) {
					return false, nil, errors.New("malformed resource")
				}
			}
			return true, package_, nil
		} else {
			return false, nil, err
		}
	}, nil
}
