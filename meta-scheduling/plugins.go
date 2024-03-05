package metascheduling

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	pluginspkg "github.com/nephio-experimental/tko/plugins"
	"github.com/nephio-experimental/tko/util"
)

const FIFOPrefix = "tko-meta-scheduling-"

type PluginInput struct {
	GRPC                    PluginInputGRPC         `yaml:"grpc"`
	LogFile                 string                  `yaml:"logFile"`
	SiteID                  string                  `yaml:"siteId"`
	SitePackage             util.Package            `yaml:"sitePackage"`
	TargetResourceIdentifer util.ResourceIdentifier `yaml:"targetResourceIdentifier"`
	Deployments             map[string]util.Package `yaml:"deployments"`
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
			Level2Protocol: self.MetaScheduling.Client.GRPCLevel2Protocol,
			Address:        self.MetaScheduling.Client.GRPCAddress,
			Port:           self.MetaScheduling.Client.GRPCPort,
		},
		LogFile:                 logFile,
		SiteID:                  self.SiteID,
		SitePackage:             self.SitePackage,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
		Deployments:             self.Deployments,
	}
}

func NewPluginScheduler(plugin client.Plugin) (SchedulerFunc, error) {
	switch plugin.Executor {
	case pluginspkg.Command:
		return NewCommandPluginScheduler(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin executor: %s", plugin.Executor)
	}
}

func NewCommandPluginScheduler(plugin client.Plugin) (SchedulerFunc, error) {
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

		if logFifo, err := executor.GetLogFIFO(FIFOPrefix, schedulingContext.Log); err == nil {
			input = schedulingContext.ToPluginInput(logFifo)
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
