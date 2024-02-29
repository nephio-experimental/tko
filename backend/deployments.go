package backend

import (
	"slices"
	"strings"

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
		Metadata:           util.CloneStringMap(self.Metadata),
		Prepared:           self.Prepared,
		Approved:           self.Approved,
	}
}

func (self *DeploymentInfo) UpdateFromResources(resources util.Resources, withMetadata bool) {
	if withMetadata {
		updateMetadataFromResources(self.Metadata, resources)
	}

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

func SortDeploymentInfos(deploymentInfos []DeploymentInfo) {
	slices.SortFunc(deploymentInfos, func(a DeploymentInfo, b DeploymentInfo) int {
		return strings.Compare(a.DeploymentID, b.DeploymentID)
	})
}

//
// Deployment
//

type Deployment struct {
	DeploymentInfo
	Resources util.Resources
}

func NewDeployment(templateId string, parentDemploymentId string, siteId string, metadata map[string]string, prepared bool, approved bool, resources util.Resources) *Deployment {
	if metadata == nil {
		metadata = make(map[string]string)
	}
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
		if metadata == nil {
			metadata = make(map[string]string)
		}
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

func (self *Deployment) Clone(withResources bool) *Deployment {
	if withResources {
		return &Deployment{
			DeploymentInfo: self.DeploymentInfo.Clone(),
			Resources:      util.CloneResources(self.Resources),
		}
	} else {
		return &Deployment{
			DeploymentInfo: self.DeploymentInfo.Clone(),
		}
	}
}

func (self *Deployment) EncodeResources(format string) ([]byte, error) {
	return util.EncodeResources(format, self.Resources)
}

func (self *Deployment) UpdateFromResources(withMetadata bool) {
	self.DeploymentInfo.UpdateFromResources(self.Resources, withMetadata)
}

func (self *Deployment) MergeTemplate(template *Template) {
	self.MergeTemplateInfo(&template.TemplateInfo)

	resources := util.CloneResources(template.Resources)
	resources = util.MergeResources(resources, self.Resources...)
	self.Resources = resources
}

func (self *Deployment) MergeDeploymentResource() {
	self.Resources = util.MergeResources(self.Resources, self.NewDeploymentResource())
}
