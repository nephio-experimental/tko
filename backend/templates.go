package backend

import (
	"slices"
	"strings"
	"time"

	"github.com/nephio-experimental/tko/util"
)

//
// TemplateInfo
//

type TemplateInfo struct {
	TemplateID    string
	Metadata      map[string]string
	Updated       time.Time
	DeploymentIDs []string
}

func (self *TemplateInfo) Clone(withDeployments bool) TemplateInfo {
	if withDeployments {
		return TemplateInfo{
			TemplateID:    self.TemplateID,
			Metadata:      util.CloneStringMap(self.Metadata),
			Updated:       self.Updated,
			DeploymentIDs: util.CloneStringSet(self.DeploymentIDs),
		}
	} else {
		return TemplateInfo{
			TemplateID: self.TemplateID,
			Metadata:   util.CloneStringMap(self.Metadata),
			Updated:    self.Updated,
		}
	}
}

func (self *TemplateInfo) UpdateFromResources(resources util.Resources) {
	updateMetadataFromResources(self.Metadata, resources)
}

func SortTemplateInfos(templateInfos []TemplateInfo) {
	slices.SortFunc(templateInfos, func(a TemplateInfo, b TemplateInfo) int {
		return strings.Compare(a.TemplateID, b.TemplateID)
	})
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
