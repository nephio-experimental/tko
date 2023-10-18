package backend

import (
	"github.com/nephio-experimental/tko/util"
)

//
// Backend
//

type Backend interface {
	Connect() error
	Release() error

	// All API errors can be BadArgumentError

	SetTemplate(template *Template) error             // error can be NotDoneError
	GetTemplate(templateId string) (*Template, error) // error can be NotFoundError
	DeleteTemplate(templateId string) error           // error can be NotDoneError, NotFoundError
	ListTemplates(templateIdPatterns []string, metadataPatterns map[string]string) ([]TemplateInfo, error)

	SetSite(site *Site) error             // error can be NotDoneError
	GetSite(siteId string) (*Site, error) // error can be NotFoundError
	DeleteSite(siteId string) error       // error can be NotDoneError, NotFoundError
	ListSites(siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]SiteInfo, error)

	SetDeployment(deployment *Deployment) error             // error can be NotDoneError
	GetDeployment(deploymentId string) (*Deployment, error) // error can be NotFoundError
	DeleteDeployment(deploymentId string) error             // error can be NotDoneError, NotFoundError
	ListDeployments(prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]DeploymentInfo, error)
	StartDeploymentModification(deploymentId string) (string, *Deployment, error)                 // error can be NotDoneError, NotFoundError, BusyError
	EndDeploymentModification(modificationToken string, resources util.Resources) (string, error) // error can be NotDoneError, NotFoundError, TimeoutError
	CancelDeploymentModification(modificationToken string) error                                  // error can be NotDoneError, NotFoundError

	SetPlugin(plugin *Plugin) error               // error can be NotDoneError
	GetPlugin(pluginId PluginID) (*Plugin, error) // error can be NotFoundError
	DeletePlugin(pluginId PluginID) error         // error can be NotDoneError, NotFoundError
	ListPlugins() ([]Plugin, error)
}
