package validating

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *ValidatingBackend) SetTemplate(context contextpkg.Context, template *backend.Template) error {
	if template.TemplateID == "" {
		return backend.NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(template.TemplateID) {
		return backend.NewBadArgumentError("invalid templateId")
	}

	if err := self.Validation.ValidatePackage(template.Package, false); err != nil {
		return backend.WrapBadArgumentError(err)
	}

	return self.Backend.SetTemplate(context, template)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) GetTemplate(context contextpkg.Context, templateId string) (*backend.Template, error) {
	if templateId == "" {
		return nil, backend.NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(templateId) {
		return nil, backend.NewBadArgumentError("invalid templateId")
	}

	return self.Backend.GetTemplate(context, templateId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) DeleteTemplate(context contextpkg.Context, templateId string) error {
	if templateId == "" {
		return backend.NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(templateId) {
		return backend.NewBadArgumentError("invalid templateId")
	}

	return self.Backend.DeleteTemplate(context, templateId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) ListTemplates(context contextpkg.Context, selectTemplates backend.SelectTemplates, window backend.Window) (util.Results[backend.TemplateInfo], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	return self.Backend.ListTemplates(context, selectTemplates, window)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) PurgeTemplates(context contextpkg.Context, selectTemplates backend.SelectTemplates) error {
	if err := self.Backend.PurgeTemplates(context, selectTemplates); err == nil {
		return nil
	} else if backend.IsNotImplementedError(err) {
		if results, err := self.Backend.ListTemplates(context, selectTemplates, backend.Window{MaxCount: -1}); err == nil {
			return ParallelDelete(context, results,
				func(templateInfo backend.TemplateInfo) string {
					return templateInfo.TemplateID
				},
				func(templateId string) error {
					return self.Backend.DeleteTemplate(context, templateId)
				},
			)
		} else {
			return err
		}
	} else {
		return err
	}
}
