package memory

import (
	contextpkg "context"
	"time"

	"github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/tliron/kutil/util"
)

type Deployment struct {
	*backend.Deployment
	CurrentModificationToken     string
	CurrentModificationTimestamp int64 // Unix microseconds
}

// ([backend.Backend] interface)
func (self *MemoryBackend) CreateDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Validate parent deployment
	if deployment.ParentDeploymentID != "" {
		if _, ok := self.deployments[deployment.ParentDeploymentID]; !ok {
			return backend.NewBadArgumentErrorf("unknown parent deployment: %s", deployment.ParentDeploymentID)
		}
	}

	// Validate template
	var template *backend.Template
	if deployment.TemplateID != "" {
		var ok bool
		if template, ok = self.templates[deployment.TemplateID]; !ok {
			return backend.NewBadArgumentErrorf("unknown template: %s", deployment.TemplateID)
		}
	}

	// Validate site
	var site *backend.Site
	if deployment.SiteID != "" {
		var ok bool
		if site, ok = self.sites[deployment.SiteID]; !ok {
			return backend.NewBadArgumentErrorf("unknown site: %s", deployment.SiteID)
		}
	}

	// Merge and associate with template
	if template != nil {
		deployment.MergeTemplate(template)
		deployment.UpdateFromPackage(true)
		template.AddDeployment(deployment.DeploymentID)
	}
	deployment.MergeDeploymentResource()

	// Associate with site
	if site != nil {
		site.AddDeployment(deployment.DeploymentID)
	}

	now := time.Now().UTC()
	deployment.DeploymentID = backend.NewID()
	deployment.Created = now
	deployment.Updated = now
	self.deployments[deployment.DeploymentID] = &Deployment{Deployment: deployment}

	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) GetDeployment(context contextpkg.Context, deploymentId string) (*backend.Deployment, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if deployment, ok := self.deployments[deploymentId]; ok {
		return deployment.Deployment.Clone(true), nil
	} else {
		return nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeleteDeployment(context contextpkg.Context, deploymentId string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if deployment, ok := self.deployments[deploymentId]; ok {
		self.deleteDeployment(context, deployment)
		return nil
	} else {
		return backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) ListDeployments(context contextpkg.Context, selectDeployments backend.SelectDeployments, window backend.Window) (util.Results[backend.DeploymentInfo], error) {
	self.lock.Lock()

	var deploymentInfos []backend.DeploymentInfo
	self.selectDeployments(context, selectDeployments, func(context contextpkg.Context, deployment *Deployment) {
		deploymentInfos = append(deploymentInfos, deployment.DeploymentInfo)
	})

	self.lock.Unlock()

	backend.SortDeploymentInfos(deploymentInfos)
	deploymentInfos = backend.ApplyWindow(deploymentInfos, window)
	return util.NewResultsSlice(deploymentInfos), nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) PurgeDeployments(context contextpkg.Context, selectDeployments backend.SelectDeployments) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.selectDeployments(context, selectDeployments, self.deleteDeployment)

	return nil
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
			deployment.CurrentModificationToken = backend.NewID()
			deployment.CurrentModificationTimestamp = time.Now().UnixMicro()
			return deployment.CurrentModificationToken, deployment.Deployment, nil
		} else {
			return "", nil, backend.NewBusyErrorf("deployment: %s", deploymentId)
		}
	} else {
		return "", nil, backend.NewNotFoundErrorf("deployment: %s", deploymentId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, package_ tkoutil.Package, validation *validationpkg.Validation) (string, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for deploymentId, deployment := range self.deployments {
		if deployment.CurrentModificationToken == modificationToken {
			if !self.hasModificationExpired(deployment) {
				deployment = &Deployment{Deployment: deployment.Clone(false)}

				originalTemplateId := deployment.TemplateID
				originalSiteId := deployment.SiteID
				deployment.Package = package_
				deployment.UpdateFromPackage(false)

				if validation != nil {
					// Complete validation when fully prepared
					if err := validation.ValidatePackage(package_, deployment.Prepared); err != nil {
						return "", err
					}
				}

				// Validate template

				var template *backend.Template
				if deployment.TemplateID != "" {
					var ok bool
					if template, ok = self.templates[deployment.TemplateID]; !ok {
						return "", backend.NewBadArgumentErrorf("unknown template: %s", deployment.TemplateID)
					}
				}

				// Validate site

				var site *backend.Site
				if deployment.SiteID != "" {
					var ok bool
					if site, ok = self.sites[deployment.SiteID]; !ok {
						return "", backend.NewBadArgumentErrorf("unknown site: %s", deployment.SiteID)
					}
				}

				// Update template assocation

				if deployment.TemplateID != originalTemplateId {
					if originalTemplateId != "" {
						if template, ok := self.templates[originalTemplateId]; ok {
							template.RemoveDeployment(originalTemplateId)
						} else {
							self.log.Warningf("missing template: %s", originalTemplateId)
						}
					}

					if template != nil {
						template.AddDeployment(deployment.DeploymentID)
					}
				}

				// Update site association

				if deployment.SiteID != originalSiteId {
					if originalSiteId != "" {
						if site, ok := self.sites[originalSiteId]; ok {
							site.RemoveDeployment(originalSiteId)
						} else {
							self.log.Warningf("missing site: %s", originalSiteId)
						}
					}

					if site != nil {
						site.AddDeployment(deployment.DeploymentID)
					}
				}

				deployment.Updated = time.Now().UTC()
				self.deployments[deploymentId] = deployment

				return deploymentId, nil
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

func (self *MemoryBackend) deleteDeployment(context contextpkg.Context, deployment *Deployment) {
	delete(self.deployments, deployment.DeploymentID)

	// Remove association from template
	if deployment.TemplateID != "" {
		if template, ok := self.templates[deployment.TemplateID]; ok {
			template.RemoveDeployment(deployment.DeploymentID)
		} else {
			self.log.Warningf("missing template: %s", deployment.TemplateID)
		}
	}

	// Remove association from site
	if deployment.SiteID != "" {
		if site, ok := self.sites[deployment.SiteID]; ok {
			site.RemoveDeployment(deployment.DeploymentID)
		} else {
			self.log.Warningf("missing site: %s", deployment.SiteID)
		}
	}

	// Remove child deployment associations
	for _, childDeployment := range self.deployments {
		if childDeployment.ParentDeploymentID == deployment.DeploymentID {
			childDeployment.ParentDeploymentID = ""
		}
	}
}

func (self *MemoryBackend) selectDeployments(context contextpkg.Context, selectDeployments backend.SelectDeployments, f func(context contextpkg.Context, deployment *Deployment)) {
	filterPrepared := (selectDeployments.Prepared != nil) && (*selectDeployments.Prepared == true)
	filterNotPrepared := (selectDeployments.Prepared != nil) && (*selectDeployments.Prepared == false)
	filterApproved := (selectDeployments.Approved != nil) && (*selectDeployments.Approved == true)
	filterNotApproved := (selectDeployments.Approved != nil) && (*selectDeployments.Approved == false)

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

		if (selectDeployments.ParentDeploymentID != nil) && (*selectDeployments.ParentDeploymentID != "") {
			if *selectDeployments.ParentDeploymentID != deployment.ParentDeploymentID {
				continue
			}
		}

		if !backend.MetadataMatchesPatterns(deployment.Metadata, selectDeployments.MetadataPatterns) {
			continue
		}

		if !backend.IDMatchesPatterns(deployment.TemplateID, selectDeployments.TemplateIDPatterns) {
			continue
		}

		if (selectDeployments.TemplateMetadataPatterns != nil) && (len(selectDeployments.TemplateMetadataPatterns) > 0) {
			if template, ok := self.templates[deployment.TemplateID]; ok {
				if !backend.MetadataMatchesPatterns(template.Metadata, selectDeployments.TemplateMetadataPatterns) {
					continue
				}
			} else {
				continue
			}
		}

		if !backend.IDMatchesPatterns(deployment.SiteID, selectDeployments.SiteIDPatterns) {
			continue
		}

		if (selectDeployments.SiteMetadataPatterns != nil) && (len(selectDeployments.SiteMetadataPatterns) > 0) {
			if site, ok := self.sites[deployment.SiteID]; ok {
				if !backend.MetadataMatchesPatterns(site.Metadata, selectDeployments.SiteMetadataPatterns) {
					continue
				}
			} else {
				continue
			}
		}

		f(context, deployment)
	}
}

func (self *MemoryBackend) hasModificationExpired(deployment *Deployment) bool {
	delta := time.Now().UnixMicro() - deployment.CurrentModificationTimestamp
	return delta > self.modificationWindow
}
