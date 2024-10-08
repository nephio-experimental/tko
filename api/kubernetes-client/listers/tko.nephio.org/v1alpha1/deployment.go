// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// DeploymentLister helps list Deployments.
// All objects returned here must be treated as read-only.
type DeploymentLister interface {
	// List lists all Deployments in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Deployment, err error)
	// Deployments returns an object that can list and get Deployments.
	Deployments(namespace string) DeploymentNamespaceLister
	DeploymentListerExpansion
}

// deploymentLister implements the DeploymentLister interface.
type deploymentLister struct {
	listers.ResourceIndexer[*v1alpha1.Deployment]
}

// NewDeploymentLister returns a new DeploymentLister.
func NewDeploymentLister(indexer cache.Indexer) DeploymentLister {
	return &deploymentLister{listers.New[*v1alpha1.Deployment](indexer, v1alpha1.Resource("deployment"))}
}

// Deployments returns an object that can list and get Deployments.
func (s *deploymentLister) Deployments(namespace string) DeploymentNamespaceLister {
	return deploymentNamespaceLister{listers.NewNamespaced[*v1alpha1.Deployment](s.ResourceIndexer, namespace)}
}

// DeploymentNamespaceLister helps list and get Deployments.
// All objects returned here must be treated as read-only.
type DeploymentNamespaceLister interface {
	// List lists all Deployments in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Deployment, err error)
	// Get retrieves the Deployment from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Deployment, error)
	DeploymentNamespaceListerExpansion
}

// deploymentNamespaceLister implements the DeploymentNamespaceLister
// interface.
type deploymentNamespaceLister struct {
	listers.ResourceIndexer[*v1alpha1.Deployment]
}
