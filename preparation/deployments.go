package preparation

import (
	contextpkg "context"
	"errors"
	"fmt"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

func (self *Preparation) PrepareDeployments() error {
	//self.Log.Notice("preparing deployments")
	false_ := false
	if deploymentInfos, err := self.Client.ListDeployments(nil, nil, nil, nil, nil, nil, &false_, nil); err == nil {
		for _, deploymentInfo := range deploymentInfos {
			context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
			self.PrepareDeployment(context, deploymentInfo)
			cancel()
		}
		return nil
	} else {
		return err
	}
}

func (self *Preparation) PrepareDeployment(context contextpkg.Context, deploymentInfo client.DeploymentInfo) {
	if deploymentInfo.Prepared {
		return
	}

	log := commonlog.NewKeyValueLogger(self.Log,
		"deployment", deploymentInfo.DeploymentID)

	log.Notice("preparing deployment",
		"template", deploymentInfo.TemplateID)
	if deployment, ok, err := self.Client.GetDeployment(deploymentInfo.DeploymentID); err == nil {
		if ok {
			if _, err := self.prepareDeployment(context, deploymentInfo.DeploymentID, deployment.Resources, log); err != nil {
				log.Error(err.Error())
			}
		} else {
			log.Info("deployment disappeared")
		}
	} else {
		log.Error(err.Error())
	}
}

func (self *Preparation) prepareDeployment(context contextpkg.Context, deploymentId string, deploymentResources util.Resources, log commonlog.Logger) (bool, error) {
	modified := false

	// Are we already fully prepared?
	if deployment, ok := util.DeploymentResourceIdentifier.GetResource(deploymentResources); ok {
		if util.IsPreparedAnnotation(deployment) {
			log.Info("already prepared")
			return false, nil
		}
	}

	todo := self.GetTODO(deploymentResources, log)
	for {
		if resourceIdentifier, ok := todo.Pop(); ok {
			if modified_, err := self.Client.ModifyDeployment(deploymentId, func(resources util.Resources) (bool, util.Resources, error) {
				// Must re-check because deployment may have been modified
				if resource, ok := resourceIdentifier.GetResource(resources); ok {
					if _, prepare := self.ShouldPrepare(resourceIdentifier, resource, nil); prepare != nil {
						preparationContext := self.NewContext(deploymentId, resources, resourceIdentifier, log)
						return prepare(context, preparationContext)
					} else {
						log.Errorf("no preparer",
							"resourceType", resourceIdentifier.GVK)
					}
				}
				return false, nil, nil
			}); err == nil {
				if modified_ {
					modified = true
				}
			} else {
				return false, err
			}
		} else {
			break
		}
	}

	// Are we fully prepared?
	if modified_, err := self.Client.ModifyDeployment(deploymentId, func(resources util.Resources) (bool, util.Resources, error) {
		if self.IsFullyPrepared(resources) {
			log.Info("fully prepared")
			if err := self.Validation.ValidateResources(resources, true); err == nil {
				if deployment, ok := util.DeploymentResourceIdentifier.GetResource(resources); ok {
					if !util.SetPreparedAnnotation(deployment, true) {
						return false, nil, errors.New("malformed Deployment resource")
					}

					// TODO: always auto approve?
					if !util.SetApprovedAnnotation(deployment, true) {
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
		if modified_ {
			modified = true
		}
	} else {
		return false, err
	}

	return modified, nil
}

func (self *Preparation) GetTODO(resources util.Resources, log commonlog.Logger) *util.ResourceIdentifiers {
	var todo util.ResourceIdentifiers
	for _, resource := range resources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if shouldPrepare, _ := self.ShouldPrepare(resourceIdentifier, resource, log); shouldPrepare {
				todo.Push(resourceIdentifier)
			}
		}
	}
	return &todo
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
