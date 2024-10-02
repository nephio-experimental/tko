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
// Runner
//

type Runner struct {
	Arguments  []string
	Properties map[string]string
	Remote     *Remote
}

func NewRunner(arguments []string, properties map[string]string) *Runner {
	self := Runner{
		Arguments:  arguments,
		Properties: properties,
	}

	self.Remote = self.NewRemote()

	return &self
}

func (self *Runner) Run(context contextpkg.Context, stdin io.Reader, command ...string) ([]byte, error) {
	if self.Remote == nil {
		return RunLocal(context, stdin, command...)
	} else {
		if kubernetesRest, err := GetKubernetesREST(); err == nil {
			return kubernetesRest.Run(context, self.Remote.KubernetesNamespace, self.Remote.KubernetesPod, self.Remote.KubernetesContainer, stdin, command...)
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

func (self *Runner) NewRemote() *Remote {
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
