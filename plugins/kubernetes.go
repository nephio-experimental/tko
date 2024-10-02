package plugins

import (
	"bytes"
	contextpkg "context"
	"errors"
	"io"
	"strings"
	"sync"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restpkg "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

//
// KubernetesREST
//

type KubernetesREST struct {
	Interface restpkg.Interface
	Config    *restpkg.Config
}

var kubernetesRest *KubernetesREST
var kubernetesRestLock sync.Mutex

func NewKubernetesREST() (*KubernetesREST, error) {
	if config, err := restpkg.InClusterConfig(); err == nil {
		if client, err := kubernetes.NewForConfig(config); err == nil {
			return &KubernetesREST{
				Interface: client.CoreV1().RESTClient(),
				Config:    config,
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func GetKubernetesREST() (*KubernetesREST, error) {
	kubernetesRestLock.Lock()
	defer kubernetesRestLock.Unlock()

	if kubernetesRest == nil {
		var err error
		if kubernetesRest, err = NewKubernetesREST(); err != nil {
			return nil, err
		}
	}

	return kubernetesRest, nil
}

func (self *KubernetesREST) Run(context contextpkg.Context, namespace string, podName string, containerName string, stdin io.Reader, command ...string) ([]byte, error) {
	if len(command) == 0 {
		return nil, errors.New("command must have at least one argument")
	}

	if strings.HasSuffix(command[0], ".py") {
		command = append([]string{"/home/tko/tko-python-env/bin/python"}, command...)
	}

	//command = append([]string{"/usr/bin/env", "DOCKER_HOST=tcp://127.0.0.1:2375"}, command...)

	log.Infof("execute Kubernetes: %s/%s/%s %s", namespace, podName, containerName, strings.Join(command, " "))

	request := self.Interface.Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(&core.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	if executor, err := remotecommand.NewSPDYExecutor(self.Config, "POST", request.URL()); err == nil {
		var stdout bytes.Buffer
		var stderr strings.Builder

		if err = executor.StreamWithContext(context, remotecommand.StreamOptions{
			Stdin:  stdin,
			Stdout: &stdout,
			Stderr: &stderr,
		}); err == nil {
			return stdout.Bytes(), nil
		} else {
			return nil, errorWithStderr(err, stderr.String())
		}
	} else {
		return nil, err
	}
}
