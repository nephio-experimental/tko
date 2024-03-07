package preparation

import (
	contextpkg "context"

	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

func (self *Preparation) GetPreparableResources(package_ tkoutil.Package, log commonlog.Logger) *tkoutil.ResourceIdentifiers {
	var preparableResources tkoutil.ResourceIdentifiers
	for _, resource := range package_ {
		if resourceIdentifier, ok := tkoutil.NewResourceIdentifierForResource(resource); ok {
			if isPreparable, _ := self.IsResourcePreparable(resourceIdentifier, resource, log); isPreparable {
				preparableResources.Push(resourceIdentifier)
			}
		}
	}
	return &preparableResources
}

func (self *Preparation) IsResourcePreparable(resourceIdentifier tkoutil.ResourceIdentifier, resource tkoutil.Resource, log commonlog.Logger) (bool, []PrepareFunc) {
	if prepareAnnotation, ok := tkoutil.GetPrepareAnnotation(resource); ok {
		if prepareAnnotation == tkoutil.PrepareAnnotationHere {
			if preparers, err := self.GetPreparers(resourceIdentifier.GVK); err == nil {
				if len(preparers) == 0 {
					if log != nil {
						log.Error("no preparers registered for trigger",
							"resourceType", resourceIdentifier.GVK)
					}
					return false, nil
				}

				if !tkoutil.IsPreparedAnnotation(resource) {
					return true, preparers
				} else if log != nil {
					log.Info("already prepared",
						"resource", resourceIdentifier)
				}
			} else if log != nil {
				log.Error(err.Error())
			}
		}
	} else if preparers, err := self.GetPreparers(resourceIdentifier.GVK); err == nil {
		// If there is no annotation but there *is* at least one preparer then we will still try to prepare (as if the annotation is "Here")
		if len(preparers) > 0 {
			if !tkoutil.IsPreparedAnnotation(resource) {
				return true, preparers
			} else if log != nil {
				log.Info("already prepared",
					"resource", resourceIdentifier)
			}
		}
	} else if log != nil {
		log.Error(err.Error())
	}

	return false, nil
}

func (self *Preparation) prepareResource(deploymentId string, resourceIdentifier tkoutil.ResourceIdentifier, log commonlog.Logger) bool {
	if modified, err := self.Client.ModifyDeployment(deploymentId, func(package_ tkoutil.Package) (bool, tkoutil.Package, error) {
		var resourceModified bool
		if resource, ok := resourceIdentifier.GetResource(package_); ok {
			log = commonlog.NewKeyValueLogger(log,
				"resource", resourceIdentifier)

			// Must re-check because deployment may have been modified since calling GetPreparableResources
			if isPreparable, preparers := self.IsResourcePreparable(resourceIdentifier, resource, nil); isPreparable {
				if len(preparers) > 0 {
					for _, prepare := range preparers {
						log.Info("preparing resource")

						preparationContext := self.NewContext(deploymentId, package_, resourceIdentifier, log)
						var preparerModified bool
						var err error

						context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
						defer cancel()

						if preparerModified, package_, err = prepare(context, preparationContext); err == nil {
							if preparerModified {
								resourceModified = true
							}
						} else {
							return false, nil, err
						}
					}
				} else {
					log.Info("no preparers for resource")
				}
			} else {
				log.Info("resource is no longer preparable")
			}
		} else {
			log.Info("resource disappeared")
		}

		if resourceModified {
			return true, package_, nil
		} else {
			return false, nil, nil
		}
	}); err == nil {
		return modified
	} else {
		log.Info(err.Error())
		return false
	}
}
