package plugins

import (
	"bytes"
	contextpkg "context"
	"errors"

	"github.com/tliron/commonlog"
	"gopkg.in/yaml.v2"
)

const Command = "command"

//
// CommandExecutor
//

type CommandExecutor struct {
	*Executor
}

func NewCommandExecutor(arguments []string, properties map[string]string) (*CommandExecutor, error) {
	if len(arguments) < 1 {
		return nil, errors.New("command executor must have at least one argument")
	}

	return &CommandExecutor{
		Executor: NewExecutor(arguments, properties),
	}, nil
}

func (self *CommandExecutor) GetLogFIFO(prefix string, log commonlog.Logger) (string, error) {
	if self.IsLocal() {
		logFifo := NewLogFIFO(prefix, log)
		if err := logFifo.Start(); err != nil {
			return "", err
		}
		return logFifo.Path, nil
	} else {
		return "", nil
	}
}

func (self *CommandExecutor) Execute(context contextpkg.Context, input any, output any) error {
	if self.Remote != nil {
		return self.ExecuteKubernetes(context, input, output)
	} else {
		return self.ExecuteLocal(context, input, output)
	}
}

func (self *CommandExecutor) ExecuteLocal(context contextpkg.Context, input any, output any) error {
	if input, err := yaml.Marshal(input); err == nil {
		if output_, err := Run(context, bytes.NewReader(input), self.Arguments...); err == nil {
			return yaml.Unmarshal(output_, output)
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *CommandExecutor) ExecuteKubernetes(context contextpkg.Context, input any, output any) error {
	return nil
}
