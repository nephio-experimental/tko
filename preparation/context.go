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
	DeploymentResources     util.Resources
	TargetResourceIdentifer util.ResourceIdentifier
}

func (self *Preparation) NewContext(deploymentId string, deploymentResources util.Resources, targetResourceIdentifer util.ResourceIdentifier, log commonlog.Logger) *Context {
	return &Context{
		Preparation:             self,
		Log:                     log,
		DeploymentID:            deploymentId,
		DeploymentResources:     deploymentResources,
		TargetResourceIdentifer: targetResourceIdentifer,
	}
}

func (self *Context) GetResource() (util.Resource, bool) {
	return self.TargetResourceIdentifer.GetResource(self.DeploymentResources)
}

func (self *Context) GetMergeResources(objectReferences []any) (bool, util.Resources, error) {
	if resources, err := util.GetReferentResources(objectReferences, self.DeploymentResources); err == nil {
		// Ensure that all mergeable resources have been prepared if they must be prepared
		for _, resource := range resources {
			if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
				if shouldPrepare, _ := self.Preparation.IsResourcePreparable(resourceIdentifier, resource, nil); shouldPrepare {
					self.Log.Info("aborting merge due to uprepared resource",
						"resource", resourceIdentifier)
					return false, nil, nil
				}
			}
		}

		return true, util.PrepareResourcesForMerge(resources), nil
	} else {
		return false, nil, err
	}
}
