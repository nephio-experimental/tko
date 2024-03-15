package backend

import (
	"slices"
	"strings"
	"time"

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
	Created            time.Time // millisecond precision
	Updated            time.Time // millisecond precision
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
		Created:            self.Created,
		Updated:            self.Updated,
		Prepared:           self.Prepared,
		Approved:           self.Approved,
	}
}

func (self *DeploymentInfo) UpdateFromPackage(package_ util.Package, withMetadata bool) {
	if withMetadata {
		updateMetadataFromPackage(self.Metadata, package_)
	}

	if deployment, ok := util.DeploymentResourceIdentifier.GetResource(package_); ok {
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
	Package util.Package
}

func NewDeploymentFromBytes(parentDemploymentId string, templateId string, siteId string, metadata map[string]string, prepared bool, approved bool, packageFormat string, package_ []byte) (*Deployment, error) {
	if package__, err := util.DecodePackage(packageFormat, package_); err == nil {
		if metadata == nil {
			metadata = make(map[string]string)
		}
		return &Deployment{
			DeploymentInfo: DeploymentInfo{
				ParentDeploymentID: parentDemploymentId,
				TemplateID:         templateId,
				SiteID:             siteId,
				Metadata:           metadata,
				Prepared:           prepared,
				Approved:           approved,
			},
			Package: package__,
		}, nil
	} else {
		return nil, err
	}
}

func (self *Deployment) Clone(withPackage bool) *Deployment {
	if withPackage {
		return &Deployment{
			DeploymentInfo: self.DeploymentInfo.Clone(),
			Package:        util.ClonePackage(self.Package),
		}
	} else {
		return &Deployment{
			DeploymentInfo: self.DeploymentInfo.Clone(),
		}
	}
}

func (self *Deployment) EncodePackage(format string) ([]byte, error) {
	return util.EncodePackage(format, self.Package)
}

func (self *Deployment) UpdateFromPackage(withMetadata bool) {
	self.DeploymentInfo.UpdateFromPackage(self.Package, withMetadata)
}

func (self *Deployment) MergeTemplate(template *Template) {
	self.MergeTemplateInfo(&template.TemplateInfo)

	package_ := util.ClonePackage(template.Package)
	package_ = util.MergePackage(package_, self.Package...)
	self.Package = package_
}

func (self *Deployment) MergeDeploymentResource() {
	self.Package = util.MergePackage(self.Package, self.NewDeploymentResource())
}

//
// SelectDeployments
//

type SelectDeployments struct {
	ParentDeploymentID       *string
	TemplateIDPatterns       []string
	TemplateMetadataPatterns map[string]string
	SiteIDPatterns           []string
	SiteMetadataPatterns     map[string]string
	MetadataPatterns         map[string]string
	Prepared                 *bool
	Approved                 *bool
}
