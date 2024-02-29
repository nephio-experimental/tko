package backend

import (
	contextpkg "context"

	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/tliron/kutil/util"
)

//
// Backend
//

type Backend interface {
	Connect(context contextpkg.Context) error
	Release(context contextpkg.Context) error

	// Owns and may change the contents of the template argument.
	// Ignores template DeploymentIDs.
	// Can return BadArgumentError, NotDoneError.
	SetTemplate(context contextpkg.Context, template *Template) error

	// Can return BadArgumentError, NotFoundError.
	GetTemplate(context contextpkg.Context, templateId string) (*Template, error)

	// Does *not* delete associated deployments, but removes associations.
	// Can return BadArgumentError, NotFoundError, NotDoneError.
	DeleteTemplate(context contextpkg.Context, templateId string) error

	// Can return BadArgumentError.
	ListTemplates(context contextpkg.Context, listTemplates ListTemplates) (util.Results[TemplateInfo], error)

	// Owns and may change the contents of the site argument.
	// Ignores site DeploymentIDs.
	// Can return BadArgumentError, NotDoneError.
	SetSite(context contextpkg.Context, site *Site) error

	// Can return BadArgumentError, NotFoundError.
	GetSite(context contextpkg.Context, siteId string) (*Site, error)

	// Does *not* delete associated deployments, but removes association.
	// Can return BadArgumentError, NotFoundError, NotDoneError.
	DeleteSite(context contextpkg.Context, siteId string) error

	// Can return BadArgumentError.
	ListSites(context contextpkg.Context, listSites ListSites) (util.Results[SiteInfo], error)

	// Owns and may change the contents of the deployment argument.
	// Can return BadArgumentError, NotDoneError.
	CreateDeployment(context contextpkg.Context, deployment *Deployment) error

	// Can return BadArgumentError, NotFoundError.
	GetDeployment(context contextpkg.Context, deploymentId string) (*Deployment, error)

	// Does *not* delete child deployments, but orphans them.
	// Can return BadArgumentError, NotFoundError, NotDoneError.
	DeleteDeployment(context contextpkg.Context, deploymentId string) error

	ListDeployments(context contextpkg.Context, listDeployments ListDeployments) (util.Results[DeploymentInfo], error)

	// Can return BadArgumentError, NotFoundError, NotDoneError, BusyError.
	StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *Deployment, error)

	// Owns and may change the contents of the resources argument.
	// May change TemplateID, SiteID, Prepared, Approved.
	// Does *not* modify Metadata, even if modified resources indicate a change.
	// If validation is not nil, should validate the modification. If the deployment is prepared, it should be complete validation.
	// Can return BadArgumentError, NotFoundError, NotDoneError, TimeoutError.
	EndDeploymentModification(context contextpkg.Context, modificationToken string, resources tkoutil.Resources, validation *validationpkg.Validation) (string, error)

	// Can return BadArgumentError, NotFoundError, NotDoneError.
	CancelDeploymentModification(context contextpkg.Context, modificationToken string) error

	// Owns and may change the contents of the plugin argument.
	// Can return BadArgumentError, NotDoneError.
	SetPlugin(context contextpkg.Context, plugin *Plugin) error

	// Can return BadArgumentError, NotFoundError.
	GetPlugin(context contextpkg.Context, pluginId PluginID) (*Plugin, error)

	// Can return BadArgumentError, NotFoundError, NotDoneError.
	DeletePlugin(context contextpkg.Context, pluginId PluginID) error

	// Can return BadArgumentError.
	ListPlugins(context contextpkg.Context, listPlugins ListPlugins) (util.Results[Plugin], error)
}

type ListTemplates struct {
	Offset             uint
	MaxCount           uint
	TemplateIDPatterns []string
	MetadataPatterns   map[string]string
}

type ListSites struct {
	Offset             uint
	MaxCount           uint
	SiteIDPatterns     []string
	TemplateIDPatterns []string
	MetadataPatterns   map[string]string
}

type ListDeployments struct {
	Offset                   uint
	MaxCount                 uint
	ParentDeploymentID       *string
	TemplateIDPatterns       []string
	TemplateMetadataPatterns map[string]string
	SiteIDPatterns           []string
	SiteMetadataPatterns     map[string]string
	MetadataPatterns         map[string]string
	Prepared                 *bool
	Approved                 *bool
}

type ListPlugins struct {
	Offset       uint
	MaxCount     uint
	Type         *string
	NamePatterns []string
	Executor     *string
	Trigger      *tkoutil.GVK
}
