package util

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tliron/go-ard"
	"gopkg.in/yaml.v3"
)

func ExecuteCommand(arguments []string, input any, output any) error {
	input_, err := yaml.Marshal(input)
	if err != nil {
		return err
	}

	pythonPath := os.Getenv("PYTHON_ENV")
	if pythonPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			pythonPath = filepath.Join(home, "tko-python-env")
		} else {
			return err
		}
	}
	pythonPath = filepath.Join(pythonPath, "bin")

	var output_ bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(arguments[0], arguments[1:]...)
	cmd.Env = append(cmd.Env, "PATH="+pythonPath)
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

func ExecuteKpt(image string, properties map[string]string, resource Resource, resources Resources) (Resources, error) {
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
