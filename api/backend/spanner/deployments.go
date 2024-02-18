package spanner

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) CreateDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetDeployment(context contextpkg.Context, deploymentId string) (*backend.Deployment, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeleteDeployment(context contextpkg.Context, deploymentId string) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListDeployments(context contextpkg.Context, listDeployments backend.ListDeployments) (util.Results[backend.DeploymentInfo], error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *backend.Deployment, error) {
	return "", nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, resources tkoutil.Resources, validation *validationpkg.Validation) (string, error) {
	return "", nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) CancelDeploymentModification(context contextpkg.Context, modificationToken string) error {
	return nil
}
