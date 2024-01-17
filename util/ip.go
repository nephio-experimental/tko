package util

import (
	"fmt"

	"github.com/tliron/kutil/util"
)

func IPStackLevel2Protocol(ipStack string) (string, error) {
	switch ipStack {
	case "dual":
		return "tcp", nil
	case "ipv6":
		return "tcp6", nil
	case "ipv4":
		return "tcp4", nil
	default:
		return "", fmt.Errorf("IP stack is not \"dual\", \"ipv6\", or \"ipv4\": %s", ipStack)
	}
}

func IPLevel2ProtocolAndAddress(ipStack string, address string) (string, string, error) {
	if level2protocol, err := IPStackLevel2Protocol(ipStack); err == nil {
		if address == "" {
			switch level2protocol {
			case "tcp4":
				address = "0.0.0.0"
			default:
				address = "::"
			}
		}

		return level2protocol, address, nil
	} else {
		return "", "", err
	}
}

func ValidateIPStack(ipStack string, name string) error {
	switch ipStack {
	case "dual", "ipv6", "ipv4":
		return nil
	default:
		return fmt.Errorf("%s is not \"dual\", \"ipv6\", or \"ipv4\": %s", name, ipStack)
	}
}

func ToReachableIPAddress(address string) (string, error) {
	if address_, zone, err := util.ToReachableIPAddress(address); err == nil {
		if zone != "" {
			address_ += "%" + zone
		}
		return address_, nil
	} else {
		return "", err
	}
}

type StartServerFunc func(level2protocol string, address string) error

func StartServer(ipStack string, address string, start StartServerFunc) error {
	switch ipStack {
	case "dual":
		if address == "" {
			// We need to bind separately for each protocol
			// See: https://github.com/golang/go/issues/9334
			if err := start("tcp6", "::"); err != nil {
				return err
			}
			return start("tcp4", "0.0.0.0")
		} else {
			return start("tcp", address)
		}

	case "ipv6":
		if address == "" {
			address = "::"
		}
		return start("tcp6", address)

	case "ipv4":
		if address == "" {
			address = "0.0.0.0"
		}
		return start("tcp4", address)

	default:
		return fmt.Errorf("unsupported IP stack: %s", ipStack)
	}
}
