package memory

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *MemoryBackend) SetTemplate(context contextpkg.Context, template *backend.Template) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Keep associated deployments
	if originalTemplate, ok := self.templates[template.TemplateID]; ok {
		template.DeploymentIDs = originalTemplate.DeploymentIDs
	}

	self.templates[template.TemplateID] = template

	return nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) GetTemplate(context contextpkg.Context, templateId string) (*backend.Template, error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if template, ok := self.templates[templateId]; ok {
		return template.Clone(true), nil
	} else {
		return nil, backend.NewNotFoundErrorf("template: %s", templateId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) DeleteTemplate(context contextpkg.Context, templateId string) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	if _, ok := self.templates[templateId]; ok {
		delete(self.templates, templateId)

		// Remove site associations
		for _, site := range self.sites {
			if site.TemplateID == templateId {
				site.TemplateID = ""
			}
		}

		// Remove deployment associations
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
func (self *MemoryBackend) ListTemplates(context contextpkg.Context, listTemplates backend.ListTemplates) (util.Results[backend.TemplateInfo], error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	var templateInfos []backend.TemplateInfo
	for _, template := range self.templates {
		if !backend.IDMatchesPatterns(template.TemplateID, listTemplates.TemplateIDPatterns) {
			continue
		}

		if !backend.MetadataMatchesPatterns(template.Metadata, listTemplates.MetadataPatterns) {
			continue
		}

		templateInfos = append(templateInfos, template.TemplateInfo)
	}

	return util.NewResultsSlice[backend.TemplateInfo](templateInfos), nil
}
