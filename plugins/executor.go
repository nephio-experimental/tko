package plugins

import (
	contextpkg "context"
	"io"
)

const (
	KubernetesNamespace = "_kubernetes.namespace" // optional (defaults to "tko")
	KubernetesPod       = "_kubernetes.pod"       // required
	KubernetesContainer = "_kubernetes.container" // optional (defaults to _kubernetes.pod)

	DefaultKubernetesNamespace = "tko"
)

//
// Executor
//

type Executor struct {
	Arguments  []string
	Properties map[string]string
	Remote     *Remote
}

func NewExecutor(arguments []string, properties map[string]string) *Executor {
	self := Executor{
		Arguments:  arguments,
		Properties: properties,
	}

	self.Remote = self.NewRemote()

	return &self
}

func (self *Executor) Execute(context contextpkg.Context, stdin io.Reader, command ...string) ([]byte, error) {
	if self.Remote == nil {
		return ExecuteLocal(context, stdin, command...)
	} else {
		if kubernetesRest, err := NewKubernetesREST(); err == nil {
			return kubernetesRest.Execute(context, self.Remote.KubernetesNamespace, self.Remote.KubernetesPod, self.Remote.KubernetesContainer, stdin, command...)
		} else {
			return nil, err
		}
	}
}

//
// Remote
//

type Remote struct {
	KubernetesNamespace string
	KubernetesPod       string
	KubernetesContainer string
}

func (self *Executor) NewRemote() *Remote {
	if self.Properties != nil {
		var remote Remote
		var ok bool
		if remote.KubernetesPod, ok = self.Properties[KubernetesPod]; ok {
			if remote.KubernetesNamespace, ok = self.Properties[KubernetesNamespace]; !ok {
				remote.KubernetesNamespace = DefaultKubernetesNamespace
			}

			if remote.KubernetesContainer, ok = self.Properties[KubernetesContainer]; !ok {
				remote.KubernetesContainer = remote.KubernetesPod
			}

			return &remote
		}
	}

	return nil
}
