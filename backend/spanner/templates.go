package spanner

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *SpannerBackend) SetTemplate(context contextpkg.Context, template *backend.Template) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) GetTemplate(context contextpkg.Context, templateId string) (*backend.Template, error) {
	return nil, nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) DeleteTemplate(context contextpkg.Context, templateId string) error {
	return nil
}

// ([backend.Backend] interface)
func (self *SpannerBackend) ListTemplates(context contextpkg.Context, listTemplates backend.ListTemplates) (util.Results[backend.TemplateInfo], error) {
	return nil, nil
}
