package backend

import (
	"github.com/nephio-experimental/tko/util"
	"github.com/segmentio/ksuid"
	"github.com/tliron/go-ard"
)

//
// DeploymentInfo
//

type DeploymentInfo struct {
	DeploymentID       string
	ParentDeploymentID string
	TemplateID         string
	SiteID             string
	Prepared           bool
}

func (self *DeploymentInfo) Update(resources util.Resources, reset bool) {
	if reset {
		self.Prepared = false
		self.TemplateID = ""
		self.SiteID = ""
	}

	if deployment, ok := util.DeploymentResourceIdentifier.GetResource(resources); ok {
		self.Prepared = util.IsPreparedAnnotation(deployment)
		spec := ard.With(deployment).Get("spec")
		if templateId, ok := spec.Get("templateId").String(); ok {
			self.TemplateID = templateId
		}
		if siteId, ok := spec.Get("siteId").String(); ok {
			self.SiteID = siteId
		}
	}
}

//
// Deployment
//

type Deployment struct {
	DeploymentInfo
	Resources util.Resources
}

func NewDeployment(templateId string, parentDemploymentId string, siteId string, prepared bool, resources util.Resources) *Deployment {
	return &Deployment{
		DeploymentInfo: DeploymentInfo{
			DeploymentID:       ksuid.New().String(),
			ParentDeploymentID: parentDemploymentId,
			TemplateID:         templateId,
			SiteID:             siteId,
			Prepared:           prepared,
		},
		Resources: resources,
	}
}

func (self *Deployment) Clone() *Deployment {
	return &Deployment{
		DeploymentInfo: DeploymentInfo{
			DeploymentID:       self.DeploymentID,
			ParentDeploymentID: self.ParentDeploymentID,
			TemplateID:         self.TemplateID,
			SiteID:             self.SiteID,
			Prepared:           self.Prepared,
		},
		Resources: cloneResources(self.Resources),
	}
}

func (self *Deployment) EncodeResources(format string) ([]byte, error) {
	return util.EncodeResources(format, self.Resources)
}

func (self *Deployment) UpdateInfo(reset bool) {
	self.DeploymentInfo.Update(self.Resources, reset)
}
