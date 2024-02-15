package preparation

import (
	contextpkg "context"
	"errors"
	"fmt"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

func (self *Preparation) PrepareDeployments() error {
	//self.Log.Notice("preparing deployments")
	false_ := false
	if deploymentInfos, err := self.Client.ListDeployments(client.ListDeployments{Prepared: &false_}); err == nil {
		return util.IterateResults(deploymentInfos, func(deploymentInfo client.DeploymentInfo) error {
			self.PrepareDeployment(deploymentInfo)
			return nil
		})
	} else {
		return err
	}
}

func (self *Preparation) PrepareDeployment(deploymentInfo client.DeploymentInfo) {
	if deploymentInfo.Prepared {
		return
	}

	log := commonlog.NewKeyValueLogger(self.Log,
		"deployment", deploymentInfo.DeploymentID)

	log.Notice("preparing deployment",
		"template", deploymentInfo.TemplateID)
	if deployment, ok, err := self.Client.GetDeployment(deploymentInfo.DeploymentID); err == nil {
		if ok {
			if _, err := self.prepareDeployment(deploymentInfo.DeploymentID, deployment.Resources, log); err != nil {
				log.Error(err.Error())
			}
		} else {
			log.Info("deployment disappeared")
		}
	} else {
		log.Error(err.Error())
	}
}

func (self *Preparation) prepareDeployment(deploymentId string, deploymentResources tkoutil.Resources, log commonlog.Logger) (bool, error) {
	deploymentModified := false

	// Are we already fully prepared?
	if deployment, ok := tkoutil.DeploymentResourceIdentifier.GetResource(deploymentResources); ok {
		if tkoutil.IsPreparedAnnotation(deployment) {
			log.Info("already prepared")
			return false, nil
		}
	}

	preparableResources := self.GetPreparableResources(deploymentResources, log)
	for {
		if resourceIdentifier, ok := preparableResources.Pop(); ok {
			if resourcesModified, err := self.Client.ModifyDeployment(deploymentId, func(resources tkoutil.Resources) (bool, tkoutil.Resources, error) {
				var resourceModified bool
				if resource, ok := resourceIdentifier.GetResource(resources); ok {
					// Must re-check because deployment may have been modified since calling GetPreparableResources
					if shouldPrepare, preparers := self.ShouldPrepare(resourceIdentifier, resource, nil); shouldPrepare {
						for _, prepare := range preparers {
							preparationContext := self.NewContext(deploymentId, resources, resourceIdentifier, log)
							var preparerModified bool
							var err error
							context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
							if preparerModified, resources, err = prepare(context, preparationContext); err == nil {
								if preparerModified {
									resourceModified = true
								}
							} else {
								cancel()
								return false, nil, err
							}
							cancel()
						}
					}
				}

				if resourceModified {
					return true, resources, nil
				} else {
					return false, nil, nil
				}
			}); err == nil {
				if resourcesModified {
					deploymentModified = true
				}
			} else {
				return false, err
			}
		} else {
			break
		}
	}

	// If we're fully prepared then update annotations
	if resourcesModified, err := self.Client.ModifyDeployment(deploymentId, func(resources tkoutil.Resources) (bool, tkoutil.Resources, error) {
		if self.IsFullyPrepared(resources) {
			log.Info("fully prepared")
			if err := self.Validation.ValidateResources(resources, true); err == nil {
				if deployment, ok := tkoutil.DeploymentResourceIdentifier.GetResource(resources); ok {
					if !tkoutil.SetPreparedAnnotation(deployment, true) {
						return false, nil, errors.New("malformed Deployment resource")
					}

					// TODO: always auto approve?

					if !tkoutil.SetApprovedAnnotation(deployment, true) {
						return false, nil, errors.New("malformed Deployment resource")
					}

					return true, resources, nil
				} else {
					return false, nil, errors.New("missing Deployment resource")
				}
			} else {
				return false, nil, fmt.Errorf("validation: %s", err.Error())
			}
		} else {
			return false, nil, nil
		}
	}); err == nil {
		if resourcesModified {
			deploymentModified = true
		}
	} else {
		return false, err
	}

	return deploymentModified, nil
}
