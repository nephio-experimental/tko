package client

import (
	contextpkg "context"
	"errors"
	"fmt"
	"io"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/util"
)

type DeploymentInfo struct {
	DeploymentID       string `json:"deploymentId" yaml:"deploymentId"`
	ParentDeploymentID string `json:"parentDeploymentId,omitempty" yaml:"parentDeploymentId,omitempty"`
	TemplateID         string `json:"templateId" yaml:"templateId"`
	SiteID             string `json:"siteId,omitempty" yaml:"siteId,omitempty"`
	Prepared           bool   `json:"prepared" yaml:"prepared"`
	Approved           bool   `json:"approved" yaml:"approved"`
}

type Deployment struct {
	DeploymentInfo
	Resources util.Resources `json:"resources" yaml:"resources"`
}

func (self *Client) CreateDeployment(parentDeploymentId string, templateId string, siteId string, prepared bool, approved bool, mergeResources util.Resources) (bool, string, string, error) {
	if mergeResources_, err := self.encodeResources(mergeResources); err == nil {
		return self.CreateDeploymentRaw(parentDeploymentId, templateId, siteId, prepared, approved, self.ResourcesFormat, mergeResources_)
	} else {
		return false, "", "", err
	}
}

func (self *Client) CreateDeploymentRaw(parentDeploymentId string, templateId string, siteId string, prepared bool, approved bool, mergeResourcesFormat string, mergeResources []byte) (bool, string, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("createDeployment")
		if response, err := apiClient.CreateDeployment(context, &api.CreateDeployment{
			ParentDeploymentId:   parentDeploymentId,
			TemplateId:           templateId,
			SiteId:               siteId,
			Prepared:             prepared,
			Approved:             approved,
			MergeResourcesFormat: mergeResourcesFormat,
			MergeResources:       mergeResources,
		}); err == nil {
			return response.Created, response.NotCreatedReason, response.DeploymentId, nil
		} else {
			return false, "", "", err
		}
	} else {
		return false, "", "", err
	}
}

func (self *Client) GetDeployment(deploymentId string) (Deployment, bool, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("getDeployment")
		if deployment, err := apiClient.GetDeployment(context, &api.GetDeployment{DeploymentId: deploymentId, PreferredResourcesFormat: self.ResourcesFormat}); err == nil {
			if resources, err := util.DecodeResources(deployment.ResourcesFormat, deployment.Resources); err == nil {
				return Deployment{
					DeploymentInfo: DeploymentInfo{
						DeploymentID: deployment.DeploymentId,
						TemplateID:   deployment.TemplateId,
						SiteID:       deployment.SiteId,
						Prepared:     deployment.Prepared,
						Approved:     deployment.Approved,
					},
					Resources: resources,
				}, true, nil
			} else {
				return Deployment{}, false, err
			}
		} else if IsNotFoundError(err) {
			return Deployment{}, false, nil
		} else {
			return Deployment{}, false, err
		}
	} else {
		return Deployment{}, false, err
	}
}

func (self *Client) DeleteDeployment(deploymentId string) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("deleteDeployment")
		if response, err := apiClient.DeleteDeployment(context, &api.DeleteDeployment{DeploymentId: deploymentId}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func (self *Client) ListDeployments(prepared *bool, approved *bool, parentDeploymentId *string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]DeploymentInfo, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("listDeployments")
		if client, err := apiClient.ListDeployments(context, &api.ListDeployments{
			Prepared:                 prepared,
			Approved:                 approved,
			ParentDeploymentId:       parentDeploymentId,
			TemplateIdPatterns:       templateIdPatterns,
			TemplateMetadataPatterns: templateMetadataPatterns,
			SiteIdPatterns:           siteIdPatterns,
			SiteMetadataPatterns:     siteMetadataPatterns,
		}); err == nil {
			var deploymentInfos []DeploymentInfo
			for {
				if response, err := client.Recv(); err == nil {
					deploymentInfos = append(deploymentInfos, DeploymentInfo{
						DeploymentID:       response.DeploymentId,
						ParentDeploymentID: response.ParentDeploymentId,
						TemplateID:         response.TemplateId,
						SiteID:             response.SiteId,
						Prepared:           response.Prepared,
						Approved:           response.Approved,
					})
				} else if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}
			return deploymentInfos, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Client) StartDeploymentModification(deploymentId string) (bool, string, string, util.Resources, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("startDeploymentModification")
		if response, err := apiClient.StartDeploymentModification(context, &api.StartDeploymentModification{DeploymentId: deploymentId}); err == nil {
			if resources, err := util.DecodeResources(response.ResourcesFormat, response.Resources); err == nil {
				return response.Started, response.NotStartedReason, response.ModificationToken, resources, nil
			} else {
				return false, "", "", nil, err
			}
		} else {
			return false, "", "", nil, err
		}
	} else {
		return false, "", "", nil, err
	}
}

func (self *Client) EndDeploymentModification(modificationToken string, resources util.Resources) (bool, string, string, error) {
	if resources_, err := self.encodeResources(resources); err == nil {
		return self.EndDeploymentModificationRaw(modificationToken, self.ResourcesFormat, resources_)
	} else {
		return false, "", "", err
	}
}

func (self *Client) EndDeploymentModificationRaw(modificationToken string, resourcesFormat string, resources []byte) (bool, string, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("endDeploymentModification")
		if response, err := apiClient.EndDeploymentModification(context, &api.EndDeploymentModification{
			ModificationToken: modificationToken,
			ResourcesFormat:   resourcesFormat,
			Resources:         resources,
		}); err == nil {
			return response.Modified, response.NotModifiedReason, response.DeploymentId, nil
		} else {
			return false, "", "", err
		}
	} else {
		return false, "", "", err
	}
}

func (self *Client) CancelDeploymentModification(modificationToken string) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("cancelDeploymentModification")
		if response, err := apiClient.CancelDeploymentModification(context, &api.CancelDeploymentModification{ModificationToken: modificationToken}); err == nil {
			return response.Cancelled, response.NotCancelledReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type ModifyDeploymentFunc func(resources util.Resources) (bool, util.Resources, error)

func (self *Client) ModifyDeployment(deploymentId string, modify ModifyDeploymentFunc) (bool, error) {
	if started, reason, modificationToken, resources, err := self.StartDeploymentModification(deploymentId); err == nil {
		if started {
			if modified, resources_, err := modify(resources); err == nil {
				if modified {
					if resources__, err := self.encodeResources(resources_); err == nil {
						if modified, reason, _, err := self.EndDeploymentModificationRaw(modificationToken, self.ResourcesFormat, resources__); err == nil {
							if modified {
								return true, nil
							} else {
								return false, fmt.Errorf("not modified: %s", reason)
							}
						} else {
							// End modification error
							if _, _, err := self.CancelDeploymentModification(modificationToken); err != nil {
								self.log.Error(err.Error())
							}
							return false, err
						}
					} else {
						// YAML encode error
						if _, _, err := self.CancelDeploymentModification(modificationToken); err != nil {
							self.log.Error(err.Error())
						}
						return false, err
					}
				} else {
					// Not modified (no error)
					if _, _, err := self.CancelDeploymentModification(modificationToken); err != nil {
						self.log.Error(err.Error())
					}
					return false, nil
				}
			} else {
				// Modify error
				if _, _, err := self.CancelDeploymentModification(modificationToken); err != nil {
					self.log.Error(err.Error())
				}
				return false, err
			}
		} else {
			return false, errors.New(reason)
		}
	} else {
		// Start modification error
		return false, err
	}
}
