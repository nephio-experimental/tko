package backend

import (
	"github.com/nephio-experimental/tko/util"
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
	Metadata           map[string]string
	Prepared           bool
	Approved           bool
}

func (self *DeploymentInfo) Clone() DeploymentInfo {
	return DeploymentInfo{
		DeploymentID:       self.DeploymentID,
		ParentDeploymentID: self.ParentDeploymentID,
		TemplateID:         self.TemplateID,
		SiteID:             self.SiteID,
		Metadata:           cloneMetadata(self.Metadata),
		Prepared:           self.Prepared,
		Approved:           self.Approved,
	}
}

func (self *DeploymentInfo) Update(resources util.Resources, reset bool) {
	if reset {
		self.TemplateID = ""
		self.SiteID = ""
		self.Prepared = false
		self.Approved = false
	}

	updateMetadata(self.Metadata, resources)

	if deployment, ok := util.DeploymentResourceIdentifier.GetResource(resources); ok {
		self.Prepared = util.IsPreparedAnnotation(deployment)
		self.Approved = util.IsApprovedAnnotation(deployment)
		spec := ard.With(deployment).Get("spec")
		if templateId, ok := spec.Get("templateId").String(); ok {
			self.TemplateID = templateId
		}
		if siteId, ok := spec.Get("siteId").String(); ok {
			self.SiteID = siteId
		}
	}
}

func (self *DeploymentInfo) MergeTemplateInfo(templateInfo *TemplateInfo) {
	// Merge metadata
	metadata := make(map[string]string)
	for key, value := range templateInfo.Metadata {
		metadata[key] = value
	}
	for key, value := range self.Metadata {
		metadata[key] = value
	}
	self.Metadata = metadata
}

func (self *DeploymentInfo) NewDeploymentResource() util.Resource {
	return util.NewDeploymentResource(self.TemplateID, self.SiteID, self.Prepared, self.Approved)
}

//
// Deployment
//

type Deployment struct {
	DeploymentInfo
	Resources util.Resources
}

func NewDeployment(templateId string, parentDemploymentId string, siteId string, metadata map[string]string, prepared bool, approved bool, resources util.Resources) *Deployment {
	return &Deployment{
		DeploymentInfo: DeploymentInfo{
			DeploymentID:       NewID(),
			ParentDeploymentID: parentDemploymentId,
			TemplateID:         templateId,
			SiteID:             siteId,
			Metadata:           metadata,
			Prepared:           prepared,
			Approved:           approved,
		},
		Resources: resources,
	}
}

func NewDeploymentFromBytes(templateId string, parentDemploymentId string, siteId string, metadata map[string]string, prepared bool, approved bool, resourcesFormat string, resources []byte) (*Deployment, error) {
	if resources, err := util.DecodeResources(resourcesFormat, resources); err == nil {
		return &Deployment{
			DeploymentInfo: DeploymentInfo{
				DeploymentID:       NewID(),
				TemplateID:         templateId,
				ParentDeploymentID: parentDemploymentId,
				SiteID:             siteId,
				Metadata:           metadata,
				Prepared:           prepared,
				Approved:           approved,
			},
			Resources: resources,
		}, nil
	} else {
		return nil, err
	}
}

func (self *Deployment) Clone() *Deployment {
	return &Deployment{
		DeploymentInfo: self.DeploymentInfo.Clone(),
		Resources:      cloneResources(self.Resources),
	}
}

func (self *Deployment) EncodeResources(format string) ([]byte, error) {
	return util.EncodeResources(format, self.Resources)
}

func (self *Deployment) Update(reset bool) {
	self.DeploymentInfo.Update(self.Resources, reset)
}

func (self *Deployment) MergeTemplate(template *Template) {
	self.MergeTemplateInfo(&template.TemplateInfo)

	// Merge our resources over template resources
	resources := util.CopyResources(template.Resources)
	resources = util.MergeResources(resources, self.Resources...)

	// Merge default Deployment resource
	resources = util.MergeResources(resources, self.NewDeploymentResource())

	self.Resources = resources
}
