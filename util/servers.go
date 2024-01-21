package util

import (
	"github.com/tliron/kutil/util"
)

type StartServerFunc func(level2protocol string, address string) error

func StartServer(ipStack util.IPStack, address string, start StartServerFunc) error {
	for _, bind := range ipStack.ServerBinds(address) {
		if err := start(bind.Level2Protocol, bind.Address); err != nil {
			return err
		}
	}
	return nil
}
