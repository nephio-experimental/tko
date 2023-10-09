package util

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tliron/go-ard"
	"gopkg.in/yaml.v3"
)

func ExecuteCommand(arguments []string, input any, output any) error {
	input_, err := yaml.Marshal(input)
	if err != nil {
		return err
	}

	var output_ bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(arguments[0], arguments[1:]...)
	cmd.Env = append(cmd.Env, "PATH=/tmp/tko-python-env/bin")
	cmd.Stdin = bytes.NewReader(input_)
	cmd.Stdout = &output_
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return withStderr(err, stderr.String())
	}

	if err := yaml.Unmarshal(output_.Bytes(), output); err != nil {
		return err
	}

	return nil
}

func ExecuteKpt(image string, properties map[string]string, resource Resource, resources []Resource) ([]Resource, error) {
	inputs := make(map[string]string)

	for name, path := range properties {
		if value, ok := ard.With(resource).GetPath(path, ".").ConvertSimilar().String(); ok {
			inputs[name] = value
		} else {
			return nil, fmt.Errorf("property not provided: %s", path)
		}
	}

	return KptFnEval(image, inputs, resources)
}

func withStderr(err error, stderr string) error {
	stderr = strings.Trim(stderr, "\r\n")
	if stderr != "" {
		return errors.New(err.Error() + "\n" + stderr)
	} else {
		return err
	}
}
