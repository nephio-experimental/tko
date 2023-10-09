package spanner

import (
	"github.com/nephio-experimental/tko/api/backend"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetTemplate(template *backend.Template) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetTemplate(templateId string) (*backend.Template, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeleteTemplate(templateId string) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListTemplates(templateIdPatterns []string, metadataPatterns map[string]string) ([]backend.TemplateInfo, error) {
	return nil, nil
}
