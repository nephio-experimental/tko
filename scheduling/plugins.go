package scheduling

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

const FIFOPrefix = "tko-scheduling-"

type PluginInput struct {
	GRPC                    PluginInputGRPC            `yaml:"grpc"`
	LogFile                 string                     `yaml:"logFile"`
	LogAddressPort          string                     `yaml:"logAddressPort"`
	SiteID                  string                     `yaml:"siteId"`
	SitePackage             tkoutil.Package            `yaml:"sitePackage"`
	TargetResourceIdentifer tkoutil.ResourceIdentifier `yaml:"targetResourceIdentifier"`
	Deployments             map[string]tkoutil.Package `yaml:"deployments"`
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
			Level2Protocol: self.Scheduling.Client.GRPCLevel2Protocol,
			Address:        self.Scheduling.Client.GRPCAddress,
			Port:           self.Scheduling.Client.GRPCPort,
		},
		LogFile:                 logFile,
		LogAddressPort:          logAddressPort,
		SiteID:                  self.SiteID,
		SitePackage:             self.SitePackage,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
		Deployments:             self.Deployments,
	}
}

func NewPluginScheduler(plugin client.Plugin, logIpStack util.IPStack, logAddress string, logPort int) (ScheduleFunc, error) {
	switch plugin.Executor {
	case pluginspkg.Command:
		return NewCommandPluginScheduler(plugin, logIpStack, logAddress, logPort)
	case pluginspkg.Ansible:
		return NewAnsiblePluginScheduler(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin executor: %s", plugin.Executor)
	}
}

func NewCommandPluginScheduler(plugin client.Plugin, logIpStack util.IPStack, logAddress string, logPort int) (ScheduleFunc, error) {
	executor, err := pluginspkg.NewCommandExecutor(plugin.Arguments, plugin.Properties)
	if err != nil {
		return nil, err
	}

	return func(context contextpkg.Context, schedulingContext *Context) error {
		schedulingContext.Log.Info("schedule via command plugin",
			"resource", schedulingContext.TargetResourceIdentifer,
			"arguments", strings.Join(plugin.Arguments, " "))

		var input PluginInput
		var output PluginOutput

		if logFile, logAddressPort, err := executor.GetLog(FIFOPrefix, logIpStack, logAddress, logPort, schedulingContext.Scheduling.Log); err == nil {
			input = schedulingContext.ToPluginInput(logFile, logAddressPort)
		} else {
			return err
		}

		if err := executor.Execute(context, input, &output); err == nil {
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

func NewAnsiblePluginScheduler(plugin client.Plugin) (ScheduleFunc, error) {
	executor, err := pluginspkg.NewAnsibleExecutor(plugin.Arguments, plugin.Properties)
	if err != nil {
		return nil, err
	}

	return func(context contextpkg.Context, schedulingContext *Context) error {
		schedulingContext.Log.Info("schedule via Ansible plugin",
			"resource", schedulingContext.TargetResourceIdentifer,
			"arguments", strings.Join(plugin.Arguments, " "))

		input := schedulingContext.ToPluginInput("", "")
		var output PluginOutput

		if err := executor.Execute(context, input, &output); err == nil {
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
