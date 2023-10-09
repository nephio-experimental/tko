package spanner

import (
	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetDeployment(deployment *backend.Deployment) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetDeployment(deploymentId string) (*backend.Deployment, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeleteDeployment(deploymentId string) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListDeployments(prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]backend.DeploymentInfo, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) StartDeploymentModification(deploymentId string) (string, *backend.Deployment, error) {
	return "", nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) EndDeploymentModification(modificationToken string, resources []util.Resource) (string, error) {
	return "", nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) CancelDeploymentModification(modificationToken string) error {
	return nil
}
