package util

import (
	"bytes"
	contextpkg "context"
	"fmt"
	"io"
	"strings"

	"github.com/nephio-experimental/tko/api/kubernetes-client/clientset/versioned/scheme"
	"gopkg.in/yaml.v3"
	core "k8s.io/api/core/v1"
	kubernetespkg "k8s.io/client-go/kubernetes"
	restpkg "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func ExecuteKubernetesCommandYAML(namespace string, podName string, containerName string, arguments []string, input any, output any) error {
	if config, err := restpkg.InClusterConfig(); err == nil {
		if kubernetesClient, err := kubernetespkg.NewForConfig(config); err == nil {
			rest := kubernetesClient.CoreV1().RESTClient()
			return executeKubernetesCommandYaml(contextpkg.TODO(), rest, config, namespace, podName, containerName, arguments, input, output)
		} else {
			return err
		}
	} else {
		return err
	}
}

func executeKubernetesCommandYaml(context contextpkg.Context, rest restpkg.Interface, config *restpkg.Config, namespace string, podName string, containerName string, command []string, input any, output any) error {
	stdin, err := yaml.Marshal(input)
	if err != nil {
		return err
	}

	var stdout, stderr bytes.Buffer
	if err := executeKubernetesCommand(context, rest, config, namespace, podName, containerName, bytes.NewReader(stdin), &stdout, &stderr, false, command...); err == nil {
		return yaml.Unmarshal(stdout.Bytes(), output)
	} else {
		return err
	}
}

func executeKubernetesCommand(context contextpkg.Context, rest restpkg.Interface, config *restpkg.Config, namespace string, podName string, containerName string, stdin io.Reader, stdout io.Writer, stderr io.Writer, tty bool, command ...string) error {
	command = append([]string{"/home/tko/tko-python-env/bin/python"}, command...)

	var stderrCapture strings.Builder
	if stderr == nil {
		// If not redirecting stderr then make sure to capture it
		stderr = &stderrCapture
	}

	execOptions := core.PodExecOptions{
		Container: containerName,
		Command:   command,
		TTY:       tty,
		Stderr:    true,
	}

	streamOptions := remotecommand.StreamOptions{
		Tty:    tty,
		Stderr: stderr,
	}

	if stdin != nil {
		execOptions.Stdin = true
		streamOptions.Stdin = stdin
	}

	if stdout != nil {
		execOptions.Stdout = true
		streamOptions.Stdout = stdout
	}

	request := rest.Post().Namespace(namespace).Resource("pods").Name(podName).SubResource("exec").VersionedParams(&execOptions, scheme.ParameterCodec)

	if executor, err := remotecommand.NewSPDYExecutor(config, "POST", request.URL()); err == nil {
		if err = executor.StreamWithContext(context, streamOptions); err == nil {
			return nil
		} else {
			return NewExecError(err, strings.TrimRight(stderrCapture.String(), "\n"))
		}
	} else {
		return err
	}
}

type ExecError struct {
	Err    error
	Stderr string
}

func NewExecError(err error, stderr string) *ExecError {
	return &ExecError{err, stderr}
}

// ([error] interface)
func (self *ExecError) Error() string {
	return fmt.Sprintf("%s\n%s", self.Err.Error(), self.Stderr)
}
