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
	SiteResources           []util.Resource
	TargetResourceIdentifer util.ResourceIdentifier
	Deployments             map[string][]util.Resource
}

func (self *Instantiation) NewContext(siteId string, siteResources []util.Resource, targetResourceIdentifer util.ResourceIdentifier, deployments map[string][]util.Resource, log commonlog.Logger) *Context {
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
