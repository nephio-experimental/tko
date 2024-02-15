package preparation

import (
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

func (self *Preparation) GetPreparableResources(resources util.Resources, log commonlog.Logger) *util.ResourceIdentifiers {
	var preparableResources util.ResourceIdentifiers
	for _, resource := range resources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if shouldPrepare, _ := self.ShouldPrepare(resourceIdentifier, resource, log); shouldPrepare {
				preparableResources.Push(resourceIdentifier)
			}
		}
	}
	return &preparableResources
}

func (self *Preparation) IsFullyPrepared(resources util.Resources) bool {
	prepared := true
	for _, resource := range resources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if resourceIdentifier == util.DeploymentResourceIdentifier {
				continue
			}

			if shouldPrepare, _ := self.ShouldPrepare(resourceIdentifier, resource, nil); shouldPrepare {
				if !util.IsPreparedAnnotation(resource) {
					prepared = false
					break
				}
			}
		}
	}
	return prepared
}
