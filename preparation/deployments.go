package preparation

import (
	"errors"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

var falseBool = false

func (self *Preparation) PrepareDeployments() error {
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
			if _, err := self.prepareDeployment(deploymentInfo.DeploymentID, deployment.Package, log); err != nil {
				log.Error(err.Error())
			}
		} else {
			log.Info("deployment disappeared")
		}
	} else {
		log.Error(err.Error())
	}
}

func (self *Preparation) IsDeploymentFullyPrepared(package_ tkoutil.Package) bool {
	prepared := true
	for _, resource := range package_ {
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

func (self *Preparation) prepareDeployment(deploymentId string, deploymentPackage tkoutil.Package, log commonlog.Logger) (bool, error) {
	deploymentModified := false

	// Are we already fully prepared?
	if deployment, ok := tkoutil.DeploymentResourceIdentifier.GetResource(deploymentPackage); ok {
		if tkoutil.IsPreparedAnnotation(deployment) {
			log.Info("already prepared")
			return false, nil
		}
	}

	preparableResources := self.GetPreparableResources(deploymentPackage, log)
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
	if packageModified, err := self.finalizeDeploymentPreparation(deploymentId, log); err == nil {
		if packageModified {
			deploymentModified = true
		}
	} else {
		return false, err
	}

	return deploymentModified, nil
}

func (self *Preparation) finalizeDeploymentPreparation(deploymentId string, log commonlog.Logger) (bool, error) {
	return self.Client.ModifyDeployment(deploymentId, func(package_ tkoutil.Package) (bool, tkoutil.Package, error) {
		if self.IsDeploymentFullyPrepared(package_) {
			log.Info("fully prepared")

			var modified bool
			if deployment, ok := tkoutil.DeploymentResourceIdentifier.GetResource(package_); ok {
				if tkoutil.SetPreparedAnnotation(deployment, true) {
					modified = true
				}

				approve := self.AutoApprove
				if approveAnnotation, ok := tkoutil.GetApproveAnnotation(deployment); ok {
					switch approveAnnotation {
					case tkoutil.ApproveAnnotationAuto:
						approve = true
					case tkoutil.ApproveAnnotationManual:
						approve = false
					}
				}

				if approve {
					if tkoutil.SetApprovedAnnotation(deployment, true) {
						modified = true
					}
				}

				return modified, package_, nil
			} else {
				return false, nil, errors.New("missing Deployment resource")
			}
		} else {
			return false, nil, nil
		}
	})
}
