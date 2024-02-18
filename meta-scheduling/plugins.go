package metascheduling

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
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
			Level2Protocol: self.MetaScheduling.Client.GRPCLevel2Protocol,
			Address:        self.MetaScheduling.Client.GRPCAddress,
			Port:           self.MetaScheduling.Client.GRPCPort,
		},
		LogFile:                 logFile,
		SiteID:                  self.SiteID,
		SiteResources:           self.SiteResources,
		TargetResourceIdentifer: self.TargetResourceIdentifer,
		Deployments:             self.Deployments,
	}
}

func NewPluginScheduler(plugin client.Plugin) (SchedulerFunc, error) {
	switch plugin.Executor {
	case "command":
		return NewCommandPluginScheduler(plugin)
	default:
		return nil, fmt.Errorf("unsupported plugin executor: %s", plugin.Executor)
	}
}

func NewCommandPluginScheduler(plugin client.Plugin) (SchedulerFunc, error) {
	if len(plugin.Arguments) < 1 {
		return nil, errors.New("plugin of type \"command\" must have at least one argument")
	}

	return func(context contextpkg.Context, schedulingContext *Context) error {
		schedulingContext.Log.Info("schedule via command plugin",
			"resource", schedulingContext.TargetResourceIdentifer,
			"arguments", strings.Join(plugin.Arguments, " "))

		logFifo := util.NewLogFIFO("tko-meta-scheduling", schedulingContext.Log)
		if err := logFifo.Start(); err != nil {
			return err
		}

		input := schedulingContext.ToPluginInput(logFifo.Path)
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
