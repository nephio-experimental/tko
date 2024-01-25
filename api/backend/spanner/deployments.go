package spanner

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
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
func (self *SpannerBackend) ListDeployments(context contextpkg.Context, prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]backend.DeploymentInfo, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *backend.Deployment, error) {
	return "", nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, resources util.Resources) (string, error) {
	return "", nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) CancelDeploymentModification(context contextpkg.Context, modificationToken string) error {
	return nil
}
