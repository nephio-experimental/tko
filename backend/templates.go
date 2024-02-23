package backend

import (
	"github.com/nephio-experimental/tko/util"
)

//
// TemplateInfo
//

type TemplateInfo struct {
	TemplateID    string
	Metadata      map[string]string
	DeploymentIDs []string
}

func (self *TemplateInfo) Clone(withDeployments bool) TemplateInfo {
	if withDeployments {
		return TemplateInfo{
			TemplateID:    self.TemplateID,
			Metadata:      util.CloneStringMap(self.Metadata),
			DeploymentIDs: util.CloneStringSet(self.DeploymentIDs),
		}
	} else {
		return TemplateInfo{
			TemplateID: self.TemplateID,
			Metadata:   util.CloneStringMap(self.Metadata),
		}
	}
}

func (self *TemplateInfo) UpdateFromResources(resources util.Resources) {
	updateMetadataFromResources(self.Metadata, resources)
}

//
// Template
//

type Template struct {
	TemplateInfo
	Resources util.Resources
}

func NewTemplateFromBytes(templateId string, metadata map[string]string, resourcesFormat string, resources []byte) (*Template, error) {
	if resources, err := util.DecodeResources(resourcesFormat, resources); err == nil {
		if metadata == nil {
			metadata = make(map[string]string)
		}
		return &Template{
			TemplateInfo: TemplateInfo{
				TemplateID: templateId,
				Metadata:   metadata,
			},
			Resources: resources,
		}, nil
	} else {
		return nil, err
	}
}

func (self *Template) Clone(withDeployments bool) *Template {
	return &Template{
		TemplateInfo: self.TemplateInfo.Clone(withDeployments),
		Resources:    util.CloneResources(self.Resources),
	}
}

func (self *Template) UpdateFromResources() {
	self.TemplateInfo.UpdateFromResources(self.Resources)
}

func (self *Template) EncodeResources(format string) ([]byte, error) {
	return util.EncodeResources(format, self.Resources)
}

func (self *Template) AddDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.AddToStringSet(self.DeploymentIDs, deploymentId)
	return ok
}

func (self *Template) RemoveDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.RemoveFromStringSet(self.DeploymentIDs, deploymentId)
	return ok
}
