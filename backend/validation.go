package backend

import (
	contextpkg "context"
	"errors"
	"regexp"

	"github.com/nephio-experimental/tko/plugins"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/tliron/kutil/util"
)

var (
	DefaultMaxCount uint = 100
	MaxMaxCount     uint = 1000

	ParallelBufferSize = 1000
	ParallelWorkers    = 10

	_ Backend = new(ValidatingBackend)
)

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

// ([fmt.Stringer] interface)
// ([Backend] interface)
func (self *ValidatingBackend) String() string {
	return self.Backend.String()
}

// ([Backend] interface)
func (self *ValidatingBackend) SetTemplate(context contextpkg.Context, template *Template) error {
	if template.TemplateID == "" {
		return NewBadArgumentError("templateId is empty")
	}
	if !IsValidID(template.TemplateID) {
		return NewBadArgumentError("invalid templateId")
	}

	if err := self.Validation.ValidatePackage(template.Package, false); err != nil {
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
func (self *ValidatingBackend) ListTemplates(context contextpkg.Context, selectTemplates SelectTemplates, window Window) (util.Results[TemplateInfo], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	return self.Backend.ListTemplates(context, selectTemplates, window)
}

// ([Backend] interface)
func (self *ValidatingBackend) PurgeTemplates(context contextpkg.Context, selectTemplates SelectTemplates) error {
	if err := self.Backend.PurgeTemplates(context, selectTemplates); err == nil {
		return nil
	} else if IsNotImplementedError(err) {
		if results, err := self.Backend.ListTemplates(context, selectTemplates, Window{MaxCount: DefaultMaxCount}); err == nil {
			return ParallelDelete(context, results,
				func(templateInfo TemplateInfo) string {
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

// ([Backend] interface)
func (self *ValidatingBackend) SetSite(context contextpkg.Context, site *Site) error {
	if site.SiteID == "" {
		return NewBadArgumentError("siteId is empty")
	}
	if !IsValidID(site.SiteID) {
		return NewBadArgumentError("invalid siteId")
	}

	if err := self.Validation.ValidatePackage(site.Package, true); err != nil {
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
func (self *ValidatingBackend) ListSites(context contextpkg.Context, selectSites SelectSites, window Window) (util.Results[SiteInfo], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	return self.Backend.ListSites(context, selectSites, window)
}

// ([Backend] interface)
func (self *ValidatingBackend) PurgeSites(context contextpkg.Context, selectSites SelectSites) error {
	if err := self.Backend.PurgeSites(context, selectSites); err == nil {
		return nil
	} else if IsNotImplementedError(err) {
		if results, err := self.Backend.ListSites(context, selectSites, Window{MaxCount: DefaultMaxCount}); err == nil {
			return ParallelDelete(context, results,
				func(siteInfo SiteInfo) string {
					return siteInfo.SiteID
				},
				func(siteId string) error {
					return self.Backend.DeleteSite(context, siteId)
				},
			)
		} else {
			return err
		}
	} else {
		return err
	}
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
	clone.UpdateFromPackage(true)
	completeValidation := clone.Prepared

	if err := self.Validation.ValidatePackage(deployment.Package, completeValidation); err != nil {
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
func (self *ValidatingBackend) ListDeployments(context contextpkg.Context, selectDeployments SelectDeployments, window Window) (util.Results[DeploymentInfo], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	return self.Backend.ListDeployments(context, selectDeployments, window)
}

// ([Backend] interface)
func (self *ValidatingBackend) PurgeDeployments(context contextpkg.Context, selectDeployments SelectDeployments) error {
	if err := self.Backend.PurgeDeployments(context, selectDeployments); err == nil {
		return nil
	} else if IsNotImplementedError(err) {
		if results, err := self.Backend.ListDeployments(context, selectDeployments, Window{MaxCount: DefaultMaxCount}); err == nil {
			return ParallelDelete(context, results,
				func(deploymentInfo DeploymentInfo) string {
					return deploymentInfo.DeploymentID
				},
				func(deploymentId string) error {
					return self.Backend.DeleteDeployment(context, deploymentId)
				},
			)
		} else {
			return err
		}
	} else {
		return err
	}
}

// ([Backend] interface)
func (self *ValidatingBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *Deployment, error) {
	if deploymentId == "" {
		return "", nil, NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.StartDeploymentModification(context, deploymentId)
}

// ([Backend] interface)
func (self *ValidatingBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, package_ tkoutil.Package, validation *validationpkg.Validation) (string, error) {
	if modificationToken == "" {
		return "", NewBadArgumentError("modificationToken is empty")
	}

	// Partial validation before calling the wrapped backend
	if err := self.Validation.ValidatePackage(package_, false); err != nil {
		return "", WrapBadArgumentError(err)
	}

	if validation == nil {
		validation = self.Validation
	}

	// It's the wrapped backend's job to validate the complete deployment
	return self.Backend.EndDeploymentModification(context, modificationToken, package_, validation)
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
	if !plugins.IsValidPluginType(plugin.Type, false) {
		return NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, plugin.Type)
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
			return NewBadArgumentError("trigger ersion is empty")
		}
		if trigger.Kind == "" {
			return NewBadArgumentError("trigger kind is empty")
		}
	}

	return self.Backend.SetPlugin(context, plugin)
}

// ([Backend] interface)
func (self *ValidatingBackend) GetPlugin(context contextpkg.Context, pluginId PluginID) (*Plugin, error) {
	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return nil, NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type)
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
	if !plugins.IsValidPluginType(pluginId.Type, false) {
		return NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, pluginId.Type)
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
func (self *ValidatingBackend) ListPlugins(context contextpkg.Context, selectPlugins SelectPlugins, window Window) (util.Results[Plugin], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	if selectPlugins.Type != nil {
		if !plugins.IsValidPluginType(*selectPlugins.Type, true) {
			return nil, NewBadArgumentErrorf("plugin type must be %s: %s", plugins.PluginTypesDescription, *selectPlugins.Type)
		}
	}

	if selectPlugins.Trigger != nil {
		// Note: plugin.Group can be empty (for default group)
		if selectPlugins.Trigger.Version == "" {
			return nil, NewBadArgumentError("trigger version is empty")
		}
		if selectPlugins.Trigger.Kind == "" {
			return nil, NewBadArgumentError("trigger kind is empty")
		}
	}

	return self.Backend.ListPlugins(context, selectPlugins, window)
}

// ([Backend] interface)
func (self *ValidatingBackend) PurgePlugins(context contextpkg.Context, selectPlugins SelectPlugins) error {
	if err := self.Backend.PurgePlugins(context, selectPlugins); err == nil {
		return nil
	} else if IsNotImplementedError(err) {
		if results, err := self.Backend.ListPlugins(context, selectPlugins, Window{MaxCount: DefaultMaxCount}); err == nil {
			return ParallelDelete(context, results,
				func(plugin Plugin) PluginID {
					return plugin.PluginID
				},
				func(pluginId PluginID) error {
					return self.Backend.DeletePlugin(context, pluginId)
				},
			)
		} else {
			return err
		}
	} else {
		return err
	}
}

// Utils

var validIdRe = regexp.MustCompile(`^[0-9A-Za-z_.\-\/:]+$`)

func IsValidID(id string) bool {
	return validIdRe.MatchString(id)
}

func ValidateWindow(window *Window) error {
	if window.MaxCount > MaxMaxCount {
		return NewBadArgumentErrorf("maxCount is too large: %d > %d", window.MaxCount, MaxMaxCount)
	}

	if window.MaxCount == 0 {
		window.MaxCount = DefaultMaxCount
	}

	return nil
}

func ParallelDelete[R any, T any](context contextpkg.Context, results util.Results[R], getTask func(result R) T, delete_ func(task T) error) error {
	deleter := util.NewParallelExecutor[T](ParallelBufferSize, func(task T) error {
		err := delete_(task)
		if IsNotFoundError(err) {
			err = nil
		}
		return err
	})

	deleter.Start(ParallelWorkers)

	if err := util.IterateResults(results, func(result R) error {
		deleter.Queue(getTask(result))
		return nil
	}); err != nil {
		deleter.Close()
		return err
	}

	return errors.Join(deleter.Wait()...)
}
