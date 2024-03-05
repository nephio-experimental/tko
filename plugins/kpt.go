package plugins

import (
	"bytes"
	contextpkg "context"
	"errors"
	"fmt"
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
	if self.Remote != nil {
		return self.ExecuteKubernetes(context, targetResourceIdentifer, package_)
	} else {
		return self.ExecuteLocal(context, targetResourceIdentifer, package_)
	}
}

func (self *KptExecutor) ExecuteLocal(context contextpkg.Context, targetResourceIdentifer util.ResourceIdentifier, package___ util.Package) (util.Package, error) {
	var targetResource util.Resource
	var ok bool
	if targetResource, ok = targetResourceIdentifer.GetResource(package___); !ok {
		// TODO: is this an error?
		return nil, errors.New("missing target resource for kpt function")
	}

	image := self.Arguments[0]
	command := []string{"kpt", "fn", "eval", "--image=" + image, "-", "--"}

	// Add kpt inputs
	resource := ard.With(targetResource).ConvertSimilar()
	for key, path := range self.Properties {
		if !strings.HasPrefix(key, "_") {
			if value, ok := resource.GetPath(path, ".").String(); ok {
				command = append(command, key+"="+value)
			} else {
				return nil, fmt.Errorf("property not provided: %s", path)
			}
		}
	}

	if input, err := util.EncodePackage("yaml", package___); err == nil {
		if output_, err := Run(context, bytes.NewReader(input), command...); err == nil {
			if resourceList, err := util.ReadPackage("yaml", bytes.NewReader(output_)); err == nil {
				if len(resourceList) == 1 {
					if package__, ok := ard.With(resourceList[0]).Get("items").ConvertSimilar().List(); ok {
						if package___, ok = util.ToMapList(package__); ok {
							return package___, nil
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

func (self *KptExecutor) ExecuteKubernetes(context contextpkg.Context, targetResourceIdentifer util.ResourceIdentifier, package_ util.Package) (util.Package, error) {
	return nil, nil
}
