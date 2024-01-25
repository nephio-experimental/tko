package backend

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/util"
)

//
// Backend
//

type Backend interface {
	Connect(context contextpkg.Context) error
	Release(context contextpkg.Context) error

	// All API errors can be BadArgumentError

	SetTemplate(context contextpkg.Context, template *Template) error             // error can be NotDoneError
	GetTemplate(context contextpkg.Context, templateId string) (*Template, error) // error can be NotFoundError
	DeleteTemplate(context contextpkg.Context, templateId string) error           // error can be NotDoneError, NotFoundError
	ListTemplates(context contextpkg.Context, templateIdPatterns []string, metadataPatterns map[string]string) ([]TemplateInfo, error)

	SetSite(context contextpkg.Context, site *Site) error             // error can be NotDoneError
	GetSite(context contextpkg.Context, siteId string) (*Site, error) // error can be NotFoundError
	DeleteSite(context contextpkg.Context, siteId string) error       // error can be NotDoneError, NotFoundError
	ListSites(context contextpkg.Context, siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]SiteInfo, error)

	SetDeployment(context contextpkg.Context, deployment *Deployment) error             // error can be NotDoneError
	GetDeployment(context contextpkg.Context, deploymentId string) (*Deployment, error) // error can be NotFoundError
	DeleteDeployment(context contextpkg.Context, deploymentId string) error             // error can be NotDoneError, NotFoundError
	ListDeployments(context contextpkg.Context, prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]DeploymentInfo, error)
	StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *Deployment, error)                 // error can be NotDoneError, NotFoundError, BusyError
	EndDeploymentModification(context contextpkg.Context, modificationToken string, resources util.Resources) (string, error) // error can be NotDoneError, NotFoundError, TimeoutError
	CancelDeploymentModification(context contextpkg.Context, modificationToken string) error                                  // error can be NotDoneError, NotFoundError

	SetPlugin(context contextpkg.Context, plugin *Plugin) error               // error can be NotDoneError
	GetPlugin(context contextpkg.Context, pluginId PluginID) (*Plugin, error) // error can be NotFoundError
	DeletePlugin(context contextpkg.Context, pluginId PluginID) error         // error can be NotDoneError, NotFoundError
	ListPlugins(context contextpkg.Context) ([]Plugin, error)
}
