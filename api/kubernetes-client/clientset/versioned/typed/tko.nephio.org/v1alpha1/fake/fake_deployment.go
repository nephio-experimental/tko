// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDeployments implements DeploymentInterface
type FakeDeployments struct {
	Fake *FakeTkoV1alpha1
	ns   string
}

var deploymentsResource = v1alpha1.SchemeGroupVersion.WithResource("deployments")

var deploymentsKind = v1alpha1.SchemeGroupVersion.WithKind("Deployment")

// Get takes name of the deployment, and returns the corresponding deployment object, and an error if there is any.
func (c *FakeDeployments) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Deployment, err error) {
	emptyResult := &v1alpha1.Deployment{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(deploymentsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.Deployment), err
}

// List takes label and field selectors, and returns the list of Deployments that match those selectors.
func (c *FakeDeployments) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DeploymentList, err error) {
	emptyResult := &v1alpha1.DeploymentList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(deploymentsResource, deploymentsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.DeploymentList{ListMeta: obj.(*v1alpha1.DeploymentList).ListMeta}
	for _, item := range obj.(*v1alpha1.DeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested deployments.
func (c *FakeDeployments) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(deploymentsResource, c.ns, opts))

}

// Create takes the representation of a deployment and creates it.  Returns the server's representation of the deployment, and an error, if there is any.
func (c *FakeDeployments) Create(ctx context.Context, deployment *v1alpha1.Deployment, opts v1.CreateOptions) (result *v1alpha1.Deployment, err error) {
	emptyResult := &v1alpha1.Deployment{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(deploymentsResource, c.ns, deployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.Deployment), err
}

// Update takes the representation of a deployment and updates it. Returns the server's representation of the deployment, and an error, if there is any.
func (c *FakeDeployments) Update(ctx context.Context, deployment *v1alpha1.Deployment, opts v1.UpdateOptions) (result *v1alpha1.Deployment, err error) {
	emptyResult := &v1alpha1.Deployment{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(deploymentsResource, c.ns, deployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.Deployment), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeDeployments) UpdateStatus(ctx context.Context, deployment *v1alpha1.Deployment, opts v1.UpdateOptions) (result *v1alpha1.Deployment, err error) {
	emptyResult := &v1alpha1.Deployment{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(deploymentsResource, "status", c.ns, deployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.Deployment), err
}

// Delete takes name of the deployment and deletes it. Returns an error if one occurs.
func (c *FakeDeployments) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(deploymentsResource, c.ns, name, opts), &v1alpha1.Deployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDeployments) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(deploymentsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.DeploymentList{})
	return err
}

// Patch applies the patch and returns the patched deployment.
func (c *FakeDeployments) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Deployment, err error) {
	emptyResult := &v1alpha1.Deployment{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(deploymentsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.Deployment), err
}
