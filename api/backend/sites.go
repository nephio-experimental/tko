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

func (self *SiteInfo) Update(resources util.Resources) {
	updateMetadata(self.Metadata, resources)
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
		SiteInfo: SiteInfo{
			SiteID:        self.SiteID,
			TemplateID:    self.TemplateID,
			Metadata:      cloneMetadata(self.Metadata),
			DeploymentIDs: util.StringSetClone(self.DeploymentIDs),
		},
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
