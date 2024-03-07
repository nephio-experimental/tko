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
	if self.Remote == nil {
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
	if inputBytes, err := yaml.Marshal(input); err == nil {
		if stdout, err := self.Executor.Execute(context, bytes.NewReader(inputBytes), self.Arguments...); err == nil {
			return yaml.Unmarshal(stdout, output)
		} else {
			return err
		}
	} else {
		return err
	}
}
