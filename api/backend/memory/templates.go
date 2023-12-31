package memory

import (
	"github.com/nephio-experimental/tko/api/backend"
)

// ([backend.Backend] interface)
func (self *MemoryBackend) SetTemplate(template *backend.Template) error {
	template = template.Clone()
	if template.Metadata == nil {
		template.Metadata = make(map[string]string)
	}

	template.Update()

	self.lock.Lock()
	defer self.lock.Unlock()

	self.templates[template.TemplateID] = template

	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) GetTemplate(templateId string) (*backend.Template, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if template, ok := self.templates[templateId]; ok {
		return template.Clone(), nil
	} else {
		return nil, backend.NewNotFoundErrorf("template: %s", templateId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeleteTemplate(templateId string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if _, ok := self.templates[templateId]; ok {
		delete(self.templates, templateId)
		for _, site := range self.sites {
			if site.TemplateID == templateId {
				site.TemplateID = ""
			}
		}
		for _, deployment := range self.deployments {
			if deployment.TemplateID == templateId {
				deployment.TemplateID = ""
			}
		}
		return nil
	} else {
		return backend.NewNotFoundErrorf("template: %s", templateId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) ListTemplates(templateIdPatterns []string, metadataPatterns map[string]string) ([]backend.TemplateInfo, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	var templateInfos []backend.TemplateInfo
	for _, template := range self.templates {
		if backend.IdMatchesPatterns(template.TemplateID, templateIdPatterns) && backend.MetadataMatchesPatterns(template.Metadata, metadataPatterns) {
			templateInfos = append(templateInfos, template.TemplateInfo)
		}
	}

	return templateInfos, nil
}
