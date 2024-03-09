package plugins

import (
	"bytes"
	contextpkg "context"
	"errors"
	"sync"

	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
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

var loggerServer *commonlog.LoggerServer
var loggerServerLock sync.Mutex

func (self *CommandExecutor) GetLog(fifoPrefix string, ipStack util.IPStack, address string, port int, log commonlog.Logger) (string, string, error) {
	if self.Remote == nil {
		loggerFifo := commonlog.NewLoggerFIFO(fifoPrefix, log, commonlog.Info)
		if err := loggerFifo.Start(); err == nil {
			loggerFifo.Log.Info("plugin log")
			return loggerFifo.Path, "", nil
		} else {
			return "", "", err
		}
	} else {
		loggerServerLock.Lock()
		defer loggerServerLock.Unlock()

		if loggerServer == nil {
			loggerServer = commonlog.NewLoggerServer(util.DualStack, address, port, log, commonlog.Info)
			if err := loggerServer.Start(); err == nil {
				util.OnExit(loggerServer.Stop)
			} else {
				return "", "", err
			}
		}

		// In dual stack IPv6 will appear first
		for _, logAddressPort := range loggerServer.ClientAddressPorts {
			if logAddress, _, found := util.SplitIPAddressPort(logAddressPort); found {
				loggerServer.Log.Info("plugin log", "address", logAddress)
				return "", logAddressPort, nil
			}
		}

		return "", "", nil
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
