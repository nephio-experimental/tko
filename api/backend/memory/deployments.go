package memory

import (
	"time"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/util"
	"github.com/segmentio/ksuid"
)

type Deployment struct {
	*backend.Deployment
	CurrentModificationToken     string
	CurrentModificationTimestamp int64 // Unix microseconds
}

// ([backend.Backend] interface)
func (self *MemoryBackend) SetDeployment(deployment *backend.Deployment) error {
	deployment = deployment.Clone()

	self.lock.Lock()
	defer self.lock.Unlock()

	if err := self.updateDeploymentInfo(deployment, false); err != nil {
		return err
	}
	self.mergeDeploymentResources(deployment)

	self.deployments[deployment.DeploymentID] = &Deployment{Deployment: deployment}
	if template, ok := self.templates[deployment.TemplateID]; ok {
		template.AddDeployment(deployment.DeploymentID)
	}
	if deployment.SiteID != "" {
		if site, ok := self.sites[deployment.SiteID]; ok {
			site.AddDeployment(deployment.DeploymentID)
		}
	}

	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) GetDeployment(deploymentId string) (*backend.Deployment, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if deployment, ok := self.deployments[deploymentId]; ok {
		return deployment.Deployment.Clone(), nil
	} else {
		return nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeleteDeployment(deploymentId string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if deployment, ok := self.deployments[deploymentId]; ok {
		self.deleteDeployment(deploymentId, deployment)
		return nil
	} else {
		return backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

func (self *MemoryBackend) deleteDeployment(deploymentId string, deployment *Deployment) {
	delete(self.deployments, deploymentId)

	if template, ok := self.templates[deployment.TemplateID]; ok {
		template.RemoveDeployment(deploymentId)
	}

	if deployment.SiteID != "" {
		if site, ok := self.sites[deployment.SiteID]; ok {
			site.RemoveDeployment(deploymentId)
		}
	}

	// Delete child deployments
	for deploymentId_, deployment_ := range self.deployments {
		if deployment_.ParentDeploymentID == deploymentId {
			self.deleteDeployment(deploymentId_, deployment_)
		}
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) ListDeployments(prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]backend.DeploymentInfo, error) {
	filterPrepared := prepared == "true"
	filterNotPrepared := prepared == "false"

	self.lock.Lock()
	defer self.lock.Unlock()

	var deploymentInfos []backend.DeploymentInfo
	for _, deployment := range self.deployments {
		if filterPrepared && !deployment.Prepared {
			continue
		}
		if filterNotPrepared && deployment.Prepared {
			continue
		}

		if parentDeploymentId != "" {
			if parentDeploymentId != deployment.ParentDeploymentID {
				continue
			}
		}

		if len(templateIdPatterns) > 0 {
			if !backend.IdMatchesPatterns(deployment.TemplateID, templateIdPatterns) {
				continue
			}
		}

		if (templateMetadataPatterns != nil) && (len(templateMetadataPatterns) > 0) {
			if template, ok := self.templates[deployment.TemplateID]; ok {
				if !backend.MetadataMatchesPatterns(template.Metadata, templateMetadataPatterns) {
					continue
				}
			} else {
				continue
			}
		}

		if len(siteIdPatterns) > 0 {
			if !backend.IdMatchesPatterns(deployment.SiteID, siteIdPatterns) {
				continue
			}
		}

		if (siteMetadataPatterns != nil) && (len(siteMetadataPatterns) > 0) {
			if site, ok := self.sites[deployment.SiteID]; ok {
				if !backend.MetadataMatchesPatterns(site.Metadata, siteMetadataPatterns) {
					continue
				}
			} else {
				continue
			}
		}

		deploymentInfos = append(deploymentInfos, deployment.DeploymentInfo)
	}

	return deploymentInfos, nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) StartDeploymentModification(deploymentId string) (string, *backend.Deployment, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if deployment, ok := self.deployments[deploymentId]; ok {
		available := deployment.CurrentModificationToken == ""
		if !available {
			available = self.hasModificationExpired(deployment)
		}

		if available {
			deployment.CurrentModificationToken = ksuid.New().String()
			deployment.CurrentModificationTimestamp = time.Now().UnixMicro()
			return deployment.CurrentModificationToken, deployment.Deployment, nil
		} else {
			// TODO: introduce a try-again loop
			return "", nil, backend.NewBusyErrorf("deployment: %s", deploymentId)
		}
	} else {
		return "", nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) EndDeploymentModification(modificationToken string, resources []util.Resource) (string, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, deployment := range self.deployments {
		if deployment.CurrentModificationToken == modificationToken {
			if !self.hasModificationExpired(deployment) {
				// TODO: clone?
				deployment.Resources = resources

				if err := self.updateDeploymentInfo(deployment.Deployment, true); err != nil {
					// TODO: undo changes
					return "", err
				}

				deployment.CurrentModificationToken = ""
				deployment.CurrentModificationTimestamp = 0

				if template, ok := self.templates[deployment.TemplateID]; ok {
					template.AddDeployment(deployment.DeploymentID)
				}
				if deployment.SiteID != "" {
					if site, ok := self.sites[deployment.SiteID]; ok {
						site.AddDeployment(deployment.DeploymentID)
					}
				}

				return deployment.DeploymentID, nil
			} else {
				deployment.CurrentModificationToken = ""
				deployment.CurrentModificationTimestamp = 0
				return "", backend.NewTimeoutErrorf("modification token: %s", modificationToken)
			}
		}
	}

	return "", backend.NewNotFoundErrorf("modification token: %s", modificationToken)
}

// ([backend.Backend] interface)
func (self *MemoryBackend) CancelDeploymentModification(modificationToken string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, deployment := range self.deployments {
		if deployment.CurrentModificationToken == modificationToken {
			deployment.CurrentModificationToken = ""
			deployment.CurrentModificationTimestamp = 0
			return nil
		}
	}

	return backend.NewNotFoundErrorf("modification token: %s", modificationToken)
}

// Utils

// Call when lock acquired
func (self *MemoryBackend) updateDeploymentInfo(deployment *backend.Deployment, reset bool) error {
	deployment.UpdateInfo(reset)
	if deployment.ParentDeploymentID != "" {
		if _, ok := self.deployments[deployment.ParentDeploymentID]; !ok {
			return backend.NewBadArgumentErrorf("unknown parent deployment: %s", deployment.ParentDeploymentID)
		}
	}
	if deployment.SiteID != "" {
		if _, ok := self.sites[deployment.SiteID]; !ok {
			return backend.NewBadArgumentErrorf("unknown site: %s", deployment.SiteID)
		}
	}
	if _, ok := self.templates[deployment.TemplateID]; !ok {
		return backend.NewBadArgumentErrorf("unknown template: %s", deployment.TemplateID)
	}
	return nil
}

// Call when lock acquired
func (self *MemoryBackend) mergeDeploymentResources(deployment *backend.Deployment) {
	if template, ok := self.templates[deployment.TemplateID]; ok {
		resources := util.CopyResources(template.Resources)

		// Merge our resources over template resources
		resources = util.MergeResources(resources, deployment.Resources)

		// Merge default Deployment resource
		resources = util.MergeResources(resources, []util.Resource{util.NewDeploymentResource(deployment.TemplateID, deployment.SiteID, deployment.Prepared)})

		deployment.Resources = resources
	}
}

func (self *MemoryBackend) hasModificationExpired(deployment *Deployment) bool {
	delta := time.Now().UnixMicro() - deployment.CurrentModificationTimestamp
	return delta > self.modificationWindow
}
