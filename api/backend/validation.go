package backend

import (
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
func (self *ValidatingBackend) Connect() error {
	return self.Backend.Connect()
}

// ([Backend] interface)
func (self *ValidatingBackend) Release() error {
	return self.Backend.Release()
}

// ([Backend] interface)
func (self *ValidatingBackend) SetTemplate(template *Template) error {
	if template.TemplateID == "" {
		return NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(template.TemplateID) {
		return NewBadArgumentError("invalid templateId")
	}
	if err := self.Validation.ValidateResources(template.Resources, false); err != nil {
		return WrapBadArgumentError(err)
	}

	return self.Backend.SetTemplate(template)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetTemplate(templateId string) (*Template, error) {
	if templateId == "" {
		return nil, NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(templateId) {
		return nil, NewBadArgumentError("invalid templateId")
	}

	return self.Backend.GetTemplate(templateId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeleteTemplate(templateId string) error {
	if templateId == "" {
		return NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(templateId) {
		return NewBadArgumentError("invalid templateId")
	}

	return self.Backend.DeleteTemplate(templateId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListTemplates(templateIdPatterns []string, metadataPatterns map[string]string) ([]TemplateInfo, error) {
	return self.Backend.ListTemplates(templateIdPatterns, metadataPatterns)
}

// ([Backend] interface)
func (self *ValidatingBackend) SetSite(site *Site) error {
	if site.SiteID == "" {
		return NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(site.SiteID) {
		return NewBadArgumentError("invalid siteId")
	}
	if err := self.Validation.ValidateResources(site.Resources, true); err != nil {
		return WrapBadArgumentError(err)
	}

	return self.Backend.SetSite(site)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetSite(siteId string) (*Site, error) {
	if siteId == "" {
		return nil, NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(siteId) {
		return nil, NewBadArgumentError("invalid siteId")
	}

	return self.Backend.GetSite(siteId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeleteSite(siteId string) error {
	if siteId == "" {
		return NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(siteId) {
		return NewBadArgumentError("invalid siteId")
	}

	return self.Backend.DeleteSite(siteId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListSites(siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]SiteInfo, error) {
	return self.Backend.ListSites(siteIdPatterns, templateIdPatterns, metadataPatterns)
}

// ([Backend] interface)
func (self *ValidatingBackend) SetDeployment(deployment *Deployment) error {
	if deployment.DeploymentID == "" {
		return NewBadArgumentError("deploymentId is empty")
	}
	if (deployment.TemplateID != "") && !IsValidID(deployment.TemplateID) {
		return NewBadArgumentError("invalid templateId")
	}

	if err := self.Validation.ValidateResources(deployment.Resources, false); err != nil {
		return WrapBadArgumentError(err)
	}

	return self.Backend.SetDeployment(deployment)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetDeployment(deploymentId string) (*Deployment, error) {
	if deploymentId == "" {
		return nil, NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.GetDeployment(deploymentId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeleteDeployment(deploymentId string) error {
	if deploymentId == "" {
		return NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.DeleteDeployment(deploymentId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListDeployments(prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]DeploymentInfo, error) {
	return self.Backend.ListDeployments(prepared, parentDeploymentId, templateIdPatterns, templateMetadataPatterns, siteIdPatterns, siteMetadataPatterns)
}

// ([Backend] interface)
func (self *ValidatingBackend) StartDeploymentModification(deploymentId string) (string, *Deployment, error) {
	if deploymentId == "" {
		return "", nil, NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.StartDeploymentModification(deploymentId)
}

// ([Backend] interface)
func (self *ValidatingBackend) EndDeploymentModification(modificationToken string, resources util.Resources) (string, error) {
	if modificationToken == "" {
		return "", NewBadArgumentError("modificationToken is empty")
	}

	if err := self.Validation.ValidateResources(resources, false); err != nil {
		return "", WrapBadArgumentError(err)
	}

	return self.Backend.EndDeploymentModification(modificationToken, resources)
}

// ([Backend] interface)
func (self *ValidatingBackend) CancelDeploymentModification(modificationToken string) error {
	if modificationToken == "" {
		return NewBadArgumentError("modificationToken is empty")
	}

	return self.Backend.CancelDeploymentModification(modificationToken)
}

// ([Backend] interface)
func (self *ValidatingBackend) SetPlugin(plugin *Plugin) error {
	switch plugin.Type {
	case "validate", "prepare", "instantiate":
	default:
		return NewBadArgumentError("type must be \"validate\", \"prepare\", or \"instantiate\"")
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

	return self.Backend.SetPlugin(plugin)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetPlugin(pluginId PluginID) (*Plugin, error) {
	switch pluginId.Type {
	case "validate", "prepare", "instantiate":
	default:
		return nil, NewBadArgumentError("type must be \"validate\", \"prepare\", or \"instantiate\"")
	}

	// Note: plugin.Group can be empty (for default group)
	if pluginId.Version == "" {
		return nil, NewBadArgumentError("version is empty")
	}
	if pluginId.Kind == "" {
		return nil, NewBadArgumentError("kind is empty")
	}

	return self.Backend.GetPlugin(pluginId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeletePlugin(pluginId PluginID) error {
	switch pluginId.Type {
	case "validate", "prepare", "instantiate":
	default:
		return NewBadArgumentError("type must be \"validate\", \"prepare\", or \"instantiate\"")
	}

	// Note: plugin.Group can be empty (for default group)
	if pluginId.Version == "" {
		return NewBadArgumentError("version is empty")
	}
	if pluginId.Kind == "" {
		return NewBadArgumentError("kind is empty")
	}

	return self.Backend.DeletePlugin(pluginId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListPlugins() ([]Plugin, error) {
	return self.Backend.ListPlugins()
}

// Utils

func IsValidID(id string) bool {
	return !strings.Contains(id, "*")
}
