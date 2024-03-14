package memory

import (
	contextpkg "context"
	"time"

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

	template.Updated = time.Now().UTC()
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

	if template, ok := self.templates[templateId]; ok {
		self.deleteTemplate(context, template)
		return nil
	} else {
		return backend.NewNotFoundErrorf("template: %s", templateId)
	}
}

// ([backend.Backend] interface)
func (self *MemoryBackend) ListTemplates(context contextpkg.Context, selectTemplates backend.SelectTemplates, window backend.Window) (util.Results[backend.TemplateInfo], error) {
	self.lock.Lock()

	var templateInfos []backend.TemplateInfo
	self.selectTemplates(context, selectTemplates, func(context contextpkg.Context, template *backend.Template) {
		templateInfos = append(templateInfos, template.TemplateInfo)
	})

	self.lock.Unlock()

	backend.SortTemplateInfos(templateInfos)
	templateInfos = backend.ApplyWindow(templateInfos, window)
	return util.NewResultsSlice(templateInfos), nil
}

// ([backend.Backend] interface)
func (self *MemoryBackend) PurgeTemplates(context contextpkg.Context, selectTemplates backend.SelectTemplates) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.selectTemplates(context, selectTemplates, self.deleteTemplate)

	return nil
}

// Utils

func (self *MemoryBackend) deleteTemplate(context contextpkg.Context, template *backend.Template) {
	delete(self.templates, template.TemplateID)

	// Remove site associations
	for _, site := range self.sites {
		if site.TemplateID == template.TemplateID {
			site.TemplateID = ""
		}
	}

	// Remove deployment associations
	for _, deployment := range self.deployments {
		if deployment.TemplateID == template.TemplateID {
			deployment.TemplateID = ""
		}
	}
}

func (self *MemoryBackend) selectTemplates(context contextpkg.Context, selectTemplates backend.SelectTemplates, f func(context contextpkg.Context, template *backend.Template)) {
	for _, template := range self.templates {
		if !backend.IDMatchesPatterns(template.TemplateID, selectTemplates.TemplateIDPatterns) {
			continue
		}

		if !backend.MetadataMatchesPatterns(template.Metadata, selectTemplates.MetadataPatterns) {
			continue
		}

		f(context, template)
	}
}
