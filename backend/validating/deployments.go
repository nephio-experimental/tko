package validating

import (
	contextpkg "context"

	"github.com/nephio-experimental/tko/backend"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/tliron/kutil/util"
)

// ([backend.Backend] interface)
func (self *ValidatingBackend) CreateDeployment(context contextpkg.Context, deployment *backend.Deployment) error {
	if (deployment.TemplateID != "") && !IsValidID(deployment.TemplateID) {
		return backend.NewBadArgumentError("invalid templateId")
	}

	if (deployment.SiteID != "") && !IsValidID(deployment.SiteID) {
		return backend.NewBadArgumentError("invalid siteId")
	}

	// Prepared deployments must be completely valid
	clone := deployment.Clone(true)
	clone.UpdateFromPackage(true)
	completeValidation := clone.Prepared

	if err := self.Validation.ValidatePackage(deployment.Package, completeValidation); err != nil {
		return backend.WrapBadArgumentError(err)
	}

	return self.Backend.CreateDeployment(context, deployment)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) GetDeployment(context contextpkg.Context, deploymentId string) (*backend.Deployment, error) {
	if deploymentId == "" {
		return nil, backend.NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.GetDeployment(context, deploymentId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) DeleteDeployment(context contextpkg.Context, deploymentId string) error {
	if deploymentId == "" {
		return backend.NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.DeleteDeployment(context, deploymentId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) ListDeployments(context contextpkg.Context, selectDeployments backend.SelectDeployments, window backend.Window) (util.Results[backend.DeploymentInfo], error) {
	if err := ValidateWindow(&window); err != nil {
		return nil, err
	}

	return self.Backend.ListDeployments(context, selectDeployments, window)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) PurgeDeployments(context contextpkg.Context, selectDeployments backend.SelectDeployments) error {
	if err := self.Backend.PurgeDeployments(context, selectDeployments); err == nil {
		return nil
	} else if backend.IsNotImplementedError(err) {
		if results, err := self.Backend.ListDeployments(context, selectDeployments, backend.Window{MaxCount: -1}); err == nil {
			return ParallelDelete(context, results,
				func(deploymentInfo backend.DeploymentInfo) string {
					return deploymentInfo.DeploymentID
				},
				func(deploymentId string) error {
					return self.Backend.DeleteDeployment(context, deploymentId)
				},
			)
		} else {
			return err
		}
	} else {
		return err
	}
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) StartDeploymentModification(context contextpkg.Context, deploymentId string) (string, *backend.Deployment, error) {
	if deploymentId == "" {
		return "", nil, backend.NewBadArgumentError("deploymentId is empty")
	}

	return self.Backend.StartDeploymentModification(context, deploymentId)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) EndDeploymentModification(context contextpkg.Context, modificationToken string, package_ tkoutil.Package, validation *validationpkg.Validation) (string, error) {
	if modificationToken == "" {
		return "", backend.NewBadArgumentError("modificationToken is empty")
	}

	// Partial validation before calling the wrapped backend
	if err := self.Validation.ValidatePackage(package_, false); err != nil {
		return "", backend.WrapBadArgumentError(err)
	}

	if validation == nil {
		validation = self.Validation
	}

	// It's the wrapped backend's job to validate the complete deployment
	return self.Backend.EndDeploymentModification(context, modificationToken, package_, validation)
}

// ([backend.Backend] interface)
func (self *ValidatingBackend) CancelDeploymentModification(context contextpkg.Context, modificationToken string) error {
	if modificationToken == "" {
		return backend.NewBadArgumentError("modificationToken is empty")
	}

	return self.Backend.CancelDeploymentModification(context, modificationToken)
}
