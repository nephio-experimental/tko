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
// TemplateInfoSliceStream
//

type TemplateInfoSliceStream struct {
	templateInfos []TemplateInfo
	length        int
	index         int
}

func NewTemplateInfoSliceStream(templateInfos []TemplateInfo) *TemplateInfoSliceStream {
	return &TemplateInfoSliceStream{
		templateInfos: templateInfos,
		length:        len(templateInfos),
	}
}

// ([TemplateInfoStream] interface)
func (self *TemplateInfoSliceStream) Next() (TemplateInfo, bool) {
	if self.index < self.length {
		templateInfo := self.templateInfos[self.index]
		self.index++
		return templateInfo, true
	} else {
		return TemplateInfo{}, false
	}
}

//
// TemplateInfoFeedStream
//

type TemplateInfoFeedStream struct {
	channel chan TemplateInfo
}

func NewTemplateInfoFeedStream(size int) *TemplateInfoFeedStream {
	return &TemplateInfoFeedStream{
		channel: make(chan TemplateInfo, size),
	}
}

// ([TemplateInfoStream] interface)
func (self *TemplateInfoFeedStream) Next() (TemplateInfo, bool) {
	templateInfo, ok := <-self.channel
	return templateInfo, ok
}

func (self *TemplateInfoFeedStream) Send(templateInfo TemplateInfo) {
	self.channel <- templateInfo
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
