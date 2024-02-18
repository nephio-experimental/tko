package preparation

import (
	"errors"
	"sync"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

var falseBool = false

func (self *Preparation) PrepareDeployments() error {
	self.preparers = sync.Map{}
	//self.Log.Notice("preparing deployments")
	if deploymentInfos, err := self.Client.ListDeployments(client.ListDeployments{Prepared: &falseBool}); err == nil {
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

func (self *Preparation) IsDeploymentFullyPrepared(resources tkoutil.Resources) bool {
	prepared := true
	for _, resource := range resources {
		if resourceIdentifier, ok := tkoutil.NewResourceIdentifierForResource(resource); ok {
			if resourceIdentifier == tkoutil.DeploymentResourceIdentifier {
				continue
			}

			if isPreparable, _ := self.IsResourcePreparable(resourceIdentifier, resource, nil); isPreparable {
				if !tkoutil.IsPreparedAnnotation(resource) {
					prepared = false
					break
				}
			}
		}
	}
	return prepared
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
			if self.prepareResource(deploymentId, resourceIdentifier, log) {
				deploymentModified = true
			}
		} else {
			break
		}
	}

	// If we're fully prepared then update annotations
	if resourcesModified, err := self.finalizeDeploymentPreparation(deploymentId, log); err == nil {
		if resourcesModified {
			deploymentModified = true
		}
	} else {
		return false, err
	}

	return deploymentModified, nil
}

func (self *Preparation) finalizeDeploymentPreparation(deploymentId string, log commonlog.Logger) (bool, error) {
	return self.Client.ModifyDeployment(deploymentId, func(resources tkoutil.Resources) (bool, tkoutil.Resources, error) {
		if self.IsDeploymentFullyPrepared(resources) {
			log.Info("fully prepared")

			var modified bool
			if deployment, ok := tkoutil.DeploymentResourceIdentifier.GetResource(resources); ok {
				if tkoutil.SetPreparedAnnotation(deployment, true) {
					modified = true
				}

				if self.AutoApprove {
					if tkoutil.SetApprovedAnnotation(deployment, true) {
						modified = true
					}
				}

				return modified, resources, nil
			} else {
				return false, nil, errors.New("missing Deployment resource")
			}
		} else {
			return false, nil, nil
		}
	})
}
