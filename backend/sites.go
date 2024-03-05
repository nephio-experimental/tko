package backend

import (
	"slices"
	"strings"
	"time"

	"github.com/nephio-experimental/tko/util"
)

//
// SiteInfo
//

type SiteInfo struct {
	SiteID        string
	TemplateID    string
	Metadata      map[string]string
	Updated       time.Time // millisecond precision
	DeploymentIDs []string
}

func (self *SiteInfo) Clone(withDeployments bool) SiteInfo {
	if withDeployments {
		return SiteInfo{
			SiteID:        self.SiteID,
			TemplateID:    self.TemplateID,
			Metadata:      util.CloneStringMap(self.Metadata),
			Updated:       self.Updated,
			DeploymentIDs: util.CloneStringList(self.DeploymentIDs),
		}
	} else {
		return SiteInfo{
			SiteID:     self.SiteID,
			TemplateID: self.TemplateID,
			Metadata:   util.CloneStringMap(self.Metadata),
			Updated:    self.Updated,
		}
	}
}

func (self *SiteInfo) UpdateFromPackage(package_ util.Package) {
	updateMetadataFromPackage(self.Metadata, package_)
}

func (self *SiteInfo) MergeTemplateInfo(templateInfo *TemplateInfo) {
	metadata := make(map[string]string)

	for key, value := range templateInfo.Metadata {
		metadata[key] = value
	}

	for key, value := range self.Metadata {
		metadata[key] = value
	}

	self.Metadata = metadata
}

func SortSiteInfos(siteInfos []SiteInfo) {
	slices.SortFunc(siteInfos, func(a SiteInfo, b SiteInfo) int {
		return strings.Compare(a.SiteID, b.SiteID)
	})
}

//
// Site
//

type Site struct {
	SiteInfo
	Package util.Package
}

func NewSiteFromBytes(siteId string, templateId string, metadata map[string]string, packageFormat string, package_ []byte) (*Site, error) {
	if package__, err := util.DecodePackage(packageFormat, package_); err == nil {
		if metadata == nil {
			metadata = make(map[string]string)
		}
		return &Site{
			SiteInfo: SiteInfo{
				SiteID:     siteId,
				TemplateID: templateId,
				Metadata:   metadata,
			},
			Package: package__,
		}, nil
	} else {
		return nil, err
	}
}

func (self *Site) Clone(withDeployments bool) *Site {
	return &Site{
		SiteInfo: self.SiteInfo.Clone(withDeployments),
		Package:  util.ClonePackage(self.Package),
	}
}

func (self *Site) UpdateFromPackage() {
	self.SiteInfo.UpdateFromPackage(self.Package)
}

func (self *Site) EncodePackage(format string) ([]byte, error) {
	return util.EncodePackage(format, self.Package)
}

func (self *Site) AddDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.AddToStringSet(self.DeploymentIDs, deploymentId)
	return ok
}

func (self *Site) RemoveDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.RemoveFromStringSet(self.DeploymentIDs, deploymentId)
	return ok
}

func (self *Site) MergeTemplate(template *Template) {
	self.MergeTemplateInfo(&template.TemplateInfo)

	package_ := util.ClonePackage(template.Package)
	package_ = util.MergePackage(package_, self.Package...)

	self.Package = package_
}
