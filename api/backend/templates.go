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

func (self *TemplateInfo) Clone() TemplateInfo {
	return TemplateInfo{
		TemplateID:    self.TemplateID,
		Metadata:      cloneMetadata(self.Metadata),
		DeploymentIDs: util.StringSetClone(self.DeploymentIDs),
	}
}

func (self *TemplateInfo) Update(resources util.Resources) {
	updateMetadata(self.Metadata, resources)
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

func (self *Template) Clone() *Template {
	return &Template{
		TemplateInfo: self.TemplateInfo.Clone(),
		Resources:    cloneResources(self.Resources),
	}
}

func (self *Template) Update() {
	self.TemplateInfo.Update(self.Resources)
}

func (self *Template) EncodeResources(format string) ([]byte, error) {
	return util.EncodeResources(format, self.Resources)
}

func (self *Template) AddDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.StringSetAdd(self.DeploymentIDs, deploymentId)
	return ok
}

func (self *Template) RemoveDeployment(deploymentId string) bool {
	var ok bool
	self.DeploymentIDs, ok = util.StringSetRemove(self.DeploymentIDs, deploymentId)
	return ok
}
