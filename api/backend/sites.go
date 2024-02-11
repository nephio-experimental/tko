package backend

import (
	"github.com/nephio-experimental/tko/util"
)

//
// SiteInfo
//

type SiteInfo struct {
	SiteID        string
	TemplateID    string
	Metadata      map[string]string
	DeploymentIDs []string
}

func (self *SiteInfo) Clone() SiteInfo {
	return SiteInfo{
		SiteID:        self.SiteID,
		TemplateID:    self.TemplateID,
		Metadata:      cloneMetadata(self.Metadata),
		DeploymentIDs: util.StringSetClone(self.DeploymentIDs),
	}
}

func (self *SiteInfo) Update(resources util.Resources) {
	updateMetadata(self.Metadata, resources)
}

func (self *SiteInfo) MergeTemplateInfo(templateInfo *TemplateInfo) {
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

//
// SiteInfoSliceStream
//

type SiteInfoSliceStream struct {
	siteInfos []SiteInfo
	length    int
	index     int
}

func NewSiteInfoSliceStream(siteInfos []SiteInfo) *SiteInfoSliceStream {
	return &SiteInfoSliceStream{
		siteInfos: siteInfos,
		length:    len(siteInfos),
	}
}

// ([SiteInfoStream] interface)
func (self *SiteInfoSliceStream) Next() (SiteInfo, bool) {
	if self.index < self.length {
		siteInfo := self.siteInfos[self.index]
		self.index++
		return siteInfo, true
	} else {
		return SiteInfo{}, false
	}
}

//
// Site
//

type Site struct {
	SiteInfo
	Resources util.Resources
}

func NewSiteFromBytes(siteId string, templateId string, metadata map[string]string, resourcesFormat string, resources []byte) (*Site, error) {
	if resources, err := util.DecodeResources(resourcesFormat, resources); err == nil {
		return &Site{
			SiteInfo: SiteInfo{
				SiteID:     siteId,
				TemplateID: templateId,
				Metadata:   metadata,
			},
			Resources: resources,
		}, nil
	} else {
		return nil, err
	}
}

func (self *Site) Clone() *Site {
	return &Site{
		SiteInfo:  self.SiteInfo.Clone(),
		Resources: cloneResources(self.Resources),
	}
}

func (self *Site) Update() {
	self.SiteInfo.Update(self.Resources)
}

func (self *Site) EncodeResources(format string) ([]byte, error) {
	return util.EncodeResources(format, self.Resources)
}

func (self *Site) AddDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.StringSetAdd(self.DeploymentIDs, deploymentId)
	return ok
}

func (self *Site) RemoveDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.StringSetRemove(self.DeploymentIDs, deploymentId)
	return ok
}

func (self *Site) MergeTemplate(template *Template) {
	self.MergeTemplateInfo(&template.TemplateInfo)

	// Merge our resources over template resources
	resources := util.CopyResources(template.Resources)
	resources = util.MergeResources(resources, self.Resources...)

	self.Resources = resources
}
