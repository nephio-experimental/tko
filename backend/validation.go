package backend

import (
	contextpkg "context"
	"regexp"

	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/tliron/kutil/util"
)

var DefaultMaxCount uint = 100
var MaxMaxCount uint = 1000

//
// ValidatingBackend
//

type ValidatingBackend struct {
	Backend    Backend
	Validation *validationpkg.Validation
}

// Wraps an existing backend with argument validation support, including
// the running of resource validation plugins.
func NewValidatingBackend(backend Backend, validation *validationpkg.Validation) *ValidatingBackend {
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
func (self *ValidatingBackend) ListTemplates(context contextpkg.Context, listTemplates ListTemplates) (util.Results[TemplateInfo], error) {
	if listTemplates.MaxCount > MaxMaxCount {
		return nil, NewBadArgumentErrorf("maxCount is too large: %d > %d", listTemplates.MaxCount, MaxMaxCount)
	}

	if listTemplates.MaxCount == 0 {
		listTemplates.MaxCount = DefaultMaxCount
	}

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
func (self *ValidatingBackend) ListSites(context contextpkg.Context, listSites ListSites) (util.Results[SiteInfo], error) {
	if listSites.MaxCount > MaxMaxCount {
		return nil, NewBadArgumentErrorf("maxCount is too large: %d > %d", listSites.MaxCount, MaxMaxCount)
	}

	if listSites.MaxCount == 0 {
		listSites.MaxCount = DefaultMaxCount
	}

	return self.Backend.ListSites(context, listSites)
}

// ([Backend] interface)
func (self *ValidatingBackend) CreateDeployment(context contextpkg.Context, deployment *Deployment) error {
	if (deployment.TemplateID != "") && !IsValidID(deployment.TemplateID) {
		return NewBadArgumentError("invalid templateId")
	}

	if (deployment.SiteID != "") && !IsValidID(deployment.SiteID) {
		return NewBadArgumentError("invalid siteId")
	}

	// Prepared deployments must be completely valid
	clone := deployment.Clone(true)
	clone.UpdateFromResources(true)
	completeValidation := clone.Prepared

	if err := self.Validation.ValidateResources(deployment.Resources, completeValidation); err != nil {
		return WrapBadArgumentError(err)
	}

	return self.Backend.CreateDeployment(context, deployment)
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
func (self *ValidatingBackend) ListDeployments(context contextpkg.Context, listDeployments ListDeployments) (util.Results[DeploymentInfo], error) {
	if listDeployments.MaxCount > MaxMaxCount {
		return nil, NewBadArgumentErrorf("maxCount is too large: %d > %d", listDeployments.MaxCount, MaxMaxCount)
	}

	if listDeployments.MaxCount == 0 {
		listDeployments.MaxCount = DefaultMaxCount
	}

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
func (self *ValidatingBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, resources tkoutil.Resources, validation *validationpkg.Validation) (string, error) {
	if modificationToken == "" {
		return "", NewBadArgumentError("modificationToken is empty")
	}

	// Partial validation before calling the wrapped backend
	if err := self.Validation.ValidateResources(resources, false); err != nil {
		return "", WrapBadArgumentError(err)
	}

	if validation == nil {
		validation = self.Validation
	}

	// It's the wrapped backend's job to validate the complete deployment
	return self.Backend.EndDeploymentModification(context, modificationToken, resources, validation)
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
	if !tkoutil.IsValidPluginType(plugin.Type, false) {
		return NewBadArgumentErrorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, plugin.Type)
	}

	if plugin.Name == "" {
		return NewBadArgumentError("name is empty")
	}
	if !IsValidID(plugin.Name) {
		return NewBadArgumentError("invalid name")
	}

	if plugin.Executor == "" {
		return NewBadArgumentError("executor is empty")
	}

	for _, trigger := range plugin.Triggers {
		// Note: plugin.Group can be empty (for default group)
		if trigger.Version == "" {
			return NewBadArgumentError("vtrigger ersion is empty")
		}
		if trigger.Kind == "" {
			return NewBadArgumentError("trigger kind is empty")
		}
	}

	return self.Backend.SetPlugin(context, plugin)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetPlugin(context contextpkg.Context, pluginId PluginID) (*Plugin, error) {
	if !tkoutil.IsValidPluginType(pluginId.Type, false) {
		return nil, NewBadArgumentErrorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, pluginId.Type)
	}

	if pluginId.Name == "" {
		return nil, NewBadArgumentError("name is empty")
	}
	if !IsValidID(pluginId.Name) {
		return nil, NewBadArgumentError("invalid name")
	}

	return self.Backend.GetPlugin(context, pluginId)
}

// ([Backend] interface)
func (self *ValidatingBackend) DeletePlugin(context contextpkg.Context, pluginId PluginID) error {
	if !tkoutil.IsValidPluginType(pluginId.Type, false) {
		return NewBadArgumentErrorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, pluginId.Type)
	}

	if pluginId.Name == "" {
		return NewBadArgumentError("name is empty")
	}
	if !IsValidID(pluginId.Name) {
		return NewBadArgumentError("invalid name")
	}

	return self.Backend.DeletePlugin(context, pluginId)
}

// ([Backend] interface)
func (self *ValidatingBackend) ListPlugins(context contextpkg.Context, listPlugins ListPlugins) (util.Results[Plugin], error) {
	if listPlugins.MaxCount > MaxMaxCount {
		return nil, NewBadArgumentErrorf("maxCount is too large: %d > %d", listPlugins.MaxCount, MaxMaxCount)
	}

	if listPlugins.MaxCount == 0 {
		listPlugins.MaxCount = DefaultMaxCount
	}

	if listPlugins.Type != nil {
		if !tkoutil.IsValidPluginType(*listPlugins.Type, true) {
			return nil, NewBadArgumentErrorf("plugin type must be %s: %s", tkoutil.PluginTypesDescription, *listPlugins.Type)
		}
	}

	if listPlugins.Trigger != nil {
		// Note: plugin.Group can be empty (for default group)
		if listPlugins.Trigger.Version == "" {
			return nil, NewBadArgumentError("trigger version is empty")
		}
		if listPlugins.Trigger.Kind == "" {
			return nil, NewBadArgumentError("trigger kind is empty")
		}
	}

	return self.Backend.ListPlugins(context, listPlugins)
}

// Utils

var validIdRe = regexp.MustCompile(`^[0-9A-Za-z_.\-\/:]+$`)

func IsValidID(id string) bool {
	return validIdRe.MatchString(id)
}
