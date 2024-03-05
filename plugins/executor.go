package plugins

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

func (self *Executor) IsLocal() bool {
	return self.Remote == nil
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
		remote.KubernetesNamespace, _ = self.Properties[KubernetesNamespace]
		remote.KubernetesPod, _ = self.Properties[KubernetesPod]
		remote.KubernetesContainer, _ = self.Properties[KubernetesContainer]

		if remote.KubernetesPod != "" {
			if remote.KubernetesNamespace == "" {
				remote.KubernetesNamespace = DefaultKubernetesNamespace
			}
			if remote.KubernetesContainer == "" {
				remote.KubernetesContainer = remote.KubernetesPod
			}

			return &remote
		}
	}

	return nil
}
