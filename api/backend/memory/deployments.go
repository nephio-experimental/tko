package memory

import (
	contextpkg "context"
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
func (self *MemoryBackend) SetDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
	deployment = deployment.Clone()
	if deployment.Metadata == nil {
		deployment.Metadata = make(map[string]string)
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	self.mergeDeployment(deployment)
	if err := self.updateDeploymentInfo(deployment, false); err != nil {
		return err
	}

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
func (self *MemoryBackend) GetDeployment(context contextpkg.Context, deploymentId string) (*backend.Deployment, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if deployment, ok := self.deployments[deploymentId]; ok {
		return deployment.Deployment.Clone(), nil
	} else {
		return nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeleteDeployment(context contextpkg.Context, deploymentId string) error {
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
func (self *MemoryBackend) ListDeployments(context contextpkg.Context, listDeployments backend.ListDeployments) ([]backend.DeploymentInfo, error) {
	filterPrepared := (listDeployments.Prepared != nil) && (*listDeployments.Prepared == true)
	filterNotPrepared := (listDeployments.Prepared != nil) && (*listDeployments.Prepared == false)
	filterApproved := (listDeployments.Approved != nil) && (*listDeployments.Approved == true)
	filterNotApproved := (listDeployments.Approved != nil) && (*listDeployments.Approved == false)

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

		if filterApproved && !deployment.Approved {
			continue
		}
		if filterNotApproved && deployment.Approved {
			continue
		}

		if (listDeployments.ParentDeploymentID != nil) && (*listDeployments.ParentDeploymentID != "") {
			if *listDeployments.ParentDeploymentID != deployment.ParentDeploymentID {
				continue
			}
		}

		if !backend.MetadataMatchesPatterns(deployment.Metadata, listDeployments.MetadataPatterns) {
			continue
		}

		if !backend.IDMatchesPatterns(deployment.TemplateID, listDeployments.TemplateIDPatterns) {
			continue
		}

		if (listDeployments.TemplateMetadataPatterns != nil) && (len(listDeployments.TemplateMetadataPatterns) > 0) {
			if template, ok := self.templates[deployment.TemplateID]; ok {
				if !backend.MetadataMatchesPatterns(template.Metadata, listDeployments.TemplateMetadataPatterns) {
					continue
				}
			} else {
				continue
			}
		}

		if !backend.IDMatchesPatterns(deployment.SiteID, listDeployments.SiteIDPatterns) {
			continue
		}

		if (listDeployments.SiteMetadataPatterns != nil) && (len(listDeployments.SiteMetadataPatterns) > 0) {
			if site, ok := self.sites[deployment.SiteID]; ok {
				if !backend.MetadataMatchesPatterns(site.Metadata, listDeployments.SiteMetadataPatterns) {
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
func (self *MemoryBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *backend.Deployment, error) {
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
func (self *MemoryBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, resources util.Resources) (string, error) {
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
func (self *MemoryBackend) CancelDeploymentModification(context contextpkg.Context, modificationToken string) error {
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
func (self *MemoryBackend) mergeDeployment(deployment *backend.Deployment) {
	if template, ok := self.templates[deployment.TemplateID]; ok {
		deployment.MergeTemplate(template)
	}
}

func (self *MemoryBackend) hasModificationExpired(deployment *Deployment) bool {
	delta := time.Now().UnixMicro() - deployment.CurrentModificationTimestamp
	return delta > self.modificationWindow
}
