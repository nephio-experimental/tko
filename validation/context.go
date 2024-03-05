package validation

import (
	"github.com/nephio-experimental/tko/util"
)

//
// Context
//

type Context struct {
	Validation              *Validation
	Package                 util.Package
	TargetResourceIdentifer util.ResourceIdentifier
	Complete                bool
}

func (self *Validation) NewContext(package_ util.Package, targetResourceIdentifer util.ResourceIdentifier, complete bool) *Context {
	return &Context{
		Validation:              self,
		Package:                 package_,
		TargetResourceIdentifer: targetResourceIdentifer,
		Complete:                complete,
	}
}

func (self *Context) GetResource() (util.Resource, bool) {
	return self.TargetResourceIdentifer.GetResource(self.Package)
}
