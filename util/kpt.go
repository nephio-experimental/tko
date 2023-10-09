package util

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/tliron/go-ard"
)

func KptFnEval(image string, inputs map[string]string, resources []Resource) ([]ard.Map, error) {
	args := []string{"fn", "eval", "--image=" + image, "-", "--"}
	for name, value := range inputs {
		args = append(args, name+"="+value)
	}

	var output bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("kpt", args...)
	resourcesYaml, err := EncodeResources("yaml", resources)
	if err != nil {
		return nil, err
	}
	cmd.Stdin = bytes.NewReader(resourcesYaml)
	cmd.Stdout = &output
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, withStderr(err, stderr.String())
	}

	resourceList, err := ReadResources("yaml", &output)
	if err != nil {
		return nil, err
	}
	if len(resourceList) != 1 {
		return nil, errors.New("kpt function did not return items")
	}
	resources_, ok := ard.With(resourceList[0]).Get("items").List()
	if !ok {
		return nil, errors.New("kpt function returned malformed list")
	}
	resources__, ok := ToMapList(resources_)
	if !ok {
		return nil, errors.New("kpt function returned a malformed item")
	}

	return resources__, nil
}
