package preparation

import (
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

//
// Context
//

type Context struct {
	Preparation             *Preparation
	Log                     commonlog.Logger
	DeploymentID            string
	DeploymentPackage       util.Package
	TargetResourceIdentifer util.ResourceIdentifier
}

func (self *Preparation) NewContext(deploymentId string, deploymentPackage util.Package, targetResourceIdentifer util.ResourceIdentifier, log commonlog.Logger) *Context {
	return &Context{
		Preparation:             self,
		Log:                     log,
		DeploymentID:            deploymentId,
		DeploymentPackage:       deploymentPackage,
		TargetResourceIdentifer: targetResourceIdentifer,
	}
}

func (self *Context) GetTargetResource() (util.Resource, bool) {
	return self.TargetResourceIdentifer.GetResource(self.DeploymentPackage)
}

func (self *Context) GetMergePackage(objectReferences []any) (bool, util.Package, error) {
	if package_, err := util.GetReferentPackage(objectReferences, self.DeploymentPackage); err == nil {
		// Ensure that all mergeable resources have been prepared if they must be prepared
		for _, resource := range package_ {
			if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
				if shouldPrepare, _ := self.Preparation.IsResourcePreparable(resourceIdentifier, resource, nil); shouldPrepare {
					self.Log.Info("aborting merge due to uprepared resource",
						"resource", resourceIdentifier)
					return false, nil, nil
				}
			}
		}

		return true, util.PreparePackageForMerge(package_), nil
	} else {
		return false, nil, err
	}
}
