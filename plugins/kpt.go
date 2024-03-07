package plugins

import (
	"bytes"
	contextpkg "context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

const Kpt = "kpt"

//
// KptExecutor
//

type KptExecutor struct {
	*Executor
}

func NewKptExecutor(arguments []string, properties map[string]string) (*KptExecutor, error) {
	if len(arguments) != 1 {
		return nil, errors.New("kpt executor must have exactly one argument")
	}

	return &KptExecutor{
		Executor: NewExecutor(arguments, properties),
	}, nil
}

func (self *KptExecutor) Execute(context contextpkg.Context, targetResourceIdentifer util.ResourceIdentifier, package_ util.Package) (util.Package, error) {
	var targetResource util.Resource
	var ok bool
	if targetResource, ok = targetResourceIdentifer.GetResource(package_); !ok {
		// TODO: is this an error?
		return nil, errors.New("missing target resource for kpt function")
	}

	kpt := "/usr/bin/kpt"
	if self.Remote != nil {
		kpt = "/opt/kpt-podman"
	}

	image := self.Arguments[0]
	command := []string{kpt, "fn", "eval", "--image=" + image, "-", "--"}

	// Add kpt inputs
	resource := ard.With(targetResource).ConvertSimilar()
	for key, path := range self.Properties {
		// Ignore internal properties
		if !strings.HasPrefix(key, "_") {
			if value, ok := resource.GetPath(path, ".").String(); ok {
				command = append(command, key+"="+value)
			} else {
				return nil, fmt.Errorf("property not provided: %s", path)
			}
		}
	}

	/*if self.Remote != nil {
		command = bashify(command...)
	}*/

	if stdin, err := util.EncodePackage("yaml", package_); err == nil {
		if stdout, err := self.Executor.Execute(context, bytes.NewReader(stdin), command...); err == nil {
			if resourceList, err := util.DecodePackage("yaml", stdout); err == nil {
				if len(resourceList) == 1 {
					if items, ok := ard.With(resourceList[0]).Get("items").ConvertSimilar().List(); ok {
						if package__, ok := util.ToMapList(items); ok {
							return package__, nil
						} else {
							return nil, errors.New("kpt function returned a malformed item")
						}
					} else {
						return nil, errors.New("kpt function returned malformed list")
					}
				} else {
					return nil, errors.New("kpt function did not return items")
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func bashify(command ...string) []string {
	command = append([]string{".", "/home/tko/.bash_profile", "&&"}, command...)
	command = []string{"/usr/bin/bash", "-c", strconv.Quote(strings.Join(command, " "))}
	//command = append([]string{"/usr/bin/env", "KPT_FN_RUNTIME=podman"}, command...)
	return command
}
