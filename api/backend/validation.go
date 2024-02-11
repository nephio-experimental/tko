package backend

import (
	contextpkg "context"
	"strings"

	"github.com/nephio-experimental/tko/util"
	"github.com/nephio-experimental/tko/validation"
)

//
// ValidatingBackend
//

type ValidatingBackend struct {
	Backend    Backend
	Validation *validation.Validation
}

// Wraps an existing backend with argument validation support, including
// the running of resource validation plugins.
func NewValidatingBackend(backend Backend, validation *validation.Validation) *ValidatingBackend {
	return &ValidatingBackend{
		Backend:    backend,
		Validation: validation,
	}
}

// ([Backend] interface)
func (self *ValidatingBackend) Connect(context contextpkg.Context) error {
	return self.Backend.Connect(context)
}

// ([Backend] interface)
func (self *ValidatingBackend) Release(context contextpkg.Context) error {
	return self.Backend.Release(context)
}

// ([Backend] interface)
func (self *ValidatingBackend) SetTemplate(context contextpkg.Context, template *Template) error {
	if template.TemplateID == "" {
		return NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(template.TemplateID) {
		return NewBadArgumentError("invalid templateId")
	}
	if err := self.Validation.ValidateResources(template.Resources, false); err != nil {
		return WrapBadArgumentError(err)
	}

	return self.Backend.SetTemplate(context, template)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetTemplate(context contextpkg.Context, templateId string) (*Template, error) {
	if templateId == "" {
		return nil, NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(templateId) {
		return nil, NewBadArgumentError("invalid templateId")
	}

	return self.Backend.GetTemplate(context, templateId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeleteTemplate(context contextpkg.Context, templateId string) error {
	if templateId == "" {
		return NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(templateId) {
		return NewBadArgumentError("invalid templateId")
	}

	return self.Backend.DeleteTemplate(context, templateId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListTemplates(context contextpkg.Context, listTemplates ListTemplates) (TemplateInfoStream, error) {
	return self.Backend.ListTemplates(context, listTemplates)
}

// ([Backend] interface)
func (self *ValidatingBackend) SetSite(context contextpkg.Context, site *Site) error {
	if site.SiteID == "" {
		return NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(site.SiteID) {
		return NewBadArgumentError("invalid siteId")
	}
	if err := self.Validation.ValidateResources(site.Resources, true); err != nil {
		return WrapBadArgumentError(err)
	}

	return self.Backend.SetSite(context, site)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetSite(context contextpkg.Context, siteId string) (*Site, error) {
	if siteId == "" {
		return nil, NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(siteId) {
		return nil, NewBadArgumentError("invalid siteId")
	}

	return self.Backend.GetSite(context, siteId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeleteSite(context contextpkg.Context, siteId string) error {
	if siteId == "" {
		return NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(siteId) {
		return NewBadArgumentError("invalid siteId")
	}

	return self.Backend.DeleteSite(context, siteId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListSites(context contextpkg.Context, listSites ListSites) (SiteInfoStream, error) {
	return self.Backend.ListSites(context, listSites)
}

// ([Backend] interface)
func (self *ValidatingBackend) SetDeployment(context contextpkg.Context, deployment *Deployment) error {
	if deployment.DeploymentID == "" {
		return NewBadArgumentError("deploymentId is empty")
	}
	if (deployment.TemplateID != "") && !IsValidID(deployment.TemplateID) {
		return NewBadArgumentError("invalid templateId")
	}

	if err := self.Validation.ValidateResources(deployment.Resources, false); err != nil {
		return WrapBadArgumentError(err)
	}

	return self.Backend.SetDeployment(context, deployment)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetDeployment(context contextpkg.Context, deploymentId string) (*Deployment, error) {
	if deploymentId == "" {
		return nil, NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.GetDeployment(context, deploymentId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeleteDeployment(context contextpkg.Context, deploymentId string) error {
	if deploymentId == "" {
		return NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.DeleteDeployment(context, deploymentId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListDeployments(context contextpkg.Context, listDeployments ListDeployments) (DeploymentInfoStream, error) {
	return self.Backend.ListDeployments(context, listDeployments)
}

// ([Backend] interface)
func (self *ValidatingBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *Deployment, error) {
	if deploymentId == "" {
		return "", nil, NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.StartDeploymentModification(context, deploymentId)
}

// ([Backend] interface)
func (self *ValidatingBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, resources util.Resources) (string, error) {
	if modificationToken == "" {
		return "", NewBadArgumentError("modificationToken is empty")
	}

	if err := self.Validation.ValidateResources(resources, false); err != nil {
		return "", WrapBadArgumentError(err)
	}

	return self.Backend.EndDeploymentModification(context, modificationToken, resources)
}

// ([Backend] interface)
func (self *ValidatingBackend) CancelDeploymentModification(context contextpkg.Context, modificationToken string) error {
	if modificationToken == "" {
		return NewBadArgumentError("modificationToken is empty")
	}

	return self.Backend.CancelDeploymentModification(context, modificationToken)
}

// ([Backend] interface)
func (self *ValidatingBackend) SetPlugin(context contextpkg.Context, plugin *Plugin) error {
	switch plugin.Type {
	case "validate", "prepare", "schedule":
	default:
		return NewBadArgumentError("type must be \"validate\", \"prepare\", or \"schedule\"")
	}

	// Note: plugin.Group can be empty (for default group)
	if plugin.Version == "" {
		return NewBadArgumentError("version is empty")
	}
	if plugin.Kind == "" {
		return NewBadArgumentError("kind is empty")
	}
	if plugin.Executor == "" {
		return NewBadArgumentError("executor is empty")
	}

	return self.Backend.SetPlugin(context, plugin)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetPlugin(context contextpkg.Context, pluginId PluginID) (*Plugin, error) {
	switch pluginId.Type {
	case "validate", "prepare", "schedule":
	default:
		return nil, NewBadArgumentError("type must be \"validate\", \"prepare\", or \"schedule\"")
	}

	// Note: plugin.Group can be empty (for default group)
	if pluginId.Version == "" {
		return nil, NewBadArgumentError("version is empty")
	}
	if pluginId.Kind == "" {
		return nil, NewBadArgumentError("kind is empty")
	}

	return self.Backend.GetPlugin(context, pluginId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeletePlugin(context contextpkg.Context, pluginId PluginID) error {
	switch pluginId.Type {
	case "validate", "prepare", "schedule":
	default:
		return NewBadArgumentError("type must be \"validate\", \"prepare\", or \"schedule\"")
	}

	// Note: plugin.Group can be empty (for default group)
	if pluginId.Version == "" {
		return NewBadArgumentError("version is empty")
	}
	if pluginId.Kind == "" {
		return NewBadArgumentError("kind is empty")
	}

	return self.Backend.DeletePlugin(context, pluginId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListPlugins(context contextpkg.Context) (PluginStream, error) {
	return self.Backend.ListPlugins(context)
}

// Utils

func IsValidID(id string) bool {
	return !strings.Contains(id, "*")
}
