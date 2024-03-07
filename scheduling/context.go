package scheduling

import (
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// Context
//

type Context struct {
	Scheduling              *Scheduling
	Log                     commonlog.Logger
	SiteID                  string
	SitePackage             util.Package
	TargetResourceIdentifer util.ResourceIdentifier
	Deployments             map[string]util.Package
}

func (self *Scheduling) NewContext(siteId string, sitePackage util.Package, targetResourceIdentifer util.ResourceIdentifier, deployments map[string]util.Package, log commonlog.Logger) *Context {
	return &Context{
		Scheduling:              self,
		Log:                     log,
		SiteID:                  siteId,
		SitePackage:             sitePackage,
		TargetResourceIdentifer: targetResourceIdentifer,
		Deployments:             deployments,
	}
}

func (self *Context) GetResource() (util.Resource, bool) {
	return self.TargetResourceIdentifer.GetResource(self.SitePackage)
}
