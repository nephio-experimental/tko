package instantiation

import (
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// Context
//

type Context struct {
	Instantiation           *Instantiation
	Log                     commonlog.Logger
	SiteID                  string
	SiteResources           util.Resources
	TargetResourceIdentifer util.ResourceIdentifier
	Deployments             map[string]util.Resources
}

func (self *Instantiation) NewContext(siteId string, siteResources util.Resources, targetResourceIdentifer util.ResourceIdentifier, deployments map[string]util.Resources, log commonlog.Logger) *Context {
	return &Context{
		Instantiation:           self,
		Log:                     log,
		SiteID:                  siteId,
		SiteResources:           siteResources,
		TargetResourceIdentifer: targetResourceIdentifer,
		Deployments:             deployments,
	}
}

func (self *Context) GetResource() (util.Resource, bool) {
	return self.TargetResourceIdentifer.GetResource(self.SiteResources)
}
