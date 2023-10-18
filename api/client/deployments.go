package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	api "github.com/nephio-experimental/tko/grpc"
	"github.com/nephio-experimental/tko/util"
)

type DeploymentInfo struct {
	DeploymentID       string `json:"deploymentId" yaml:"deploymentId"`
	ParentDeploymentID string `json:"parentDeploymentId,omitempty" yaml:"parentDeploymentId,omitempty"`
	TemplateID         string `json:"templateId" yaml:"templateId"`
	SiteID             string `json:"siteId,omitempty" yaml:"siteId,omitempty"`
	Prepared           bool   `json:"prepared" yaml:"prepared"`
}

type Deployment struct {
	DeploymentInfo
	Resources util.Resources `json:"resources" yaml:"resources"`
}

func (self *Client) CreateDeployment(parentDeploymentId string, templateId string, siteId string, prepared bool, mergeResources util.Resources) (bool, string, string, error) {
	if mergeResources_, err := self.encodeResources(mergeResources); err == nil {
		return self.CreateDeploymentRaw(parentDeploymentId, templateId, siteId, prepared, self.ResourcesFormat, mergeResources_)
	} else {
		return false, "", "", err
	}
}

func (self *Client) CreateDeploymentRaw(parentDeploymentId string, templateId string, siteId string, prepared bool, mergeResourcesFormat string, mergeResources []byte) (bool, string, string, error) {
	if response, err := self.client.CreateDeployment(context.TODO(), &api.CreateDeployment{
		ParentDeploymentId:   parentDeploymentId,
		TemplateId:           templateId,
		SiteId:               siteId,
		Prepared:             prepared,
		MergeResourcesFormat: mergeResourcesFormat,
		MergeResources:       mergeResources,
	}); err == nil {
		return response.Created, response.NotCreatedReason, response.DeploymentId, nil
	} else {
		return false, "", "", err
	}
}

func (self *Client) GetDeployment(deploymentId string) (Deployment, bool, error) {
	if deployment, err := self.client.GetDeployment(context.TODO(), &api.GetDeployment{DeploymentId: deploymentId, PreferredResourcesFormat: self.ResourcesFormat}); err == nil {
		if resources, err := util.DecodeResources(deployment.ResourcesFormat, deployment.Resources); err == nil {
			return Deployment{
				DeploymentInfo: DeploymentInfo{
					DeploymentID: deployment.DeploymentId,
					TemplateID:   deployment.TemplateId,
					SiteID:       deployment.SiteId,
					Prepared:     deployment.Prepared,
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
}

func (self *Client) DeleteDeployment(deploymentId string) (bool, string, error) {
	if response, err := self.client.DeleteDeployment(context.TODO(), &api.DeleteDeployment{DeploymentId: deploymentId}); err == nil {
		return response.Deleted, response.NotDeletedReason, nil
	} else {
		return false, "", err
	}
}

func (self *Client) ListDeployments(prepared string, parentDeploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) ([]DeploymentInfo, error) {
	if client, err := self.client.ListDeployments(context.TODO(), &api.ListDeployments{
		Prepared:                 prepared,
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
}

func (self *Client) StartDeploymentModification(deploymentId string) (bool, string, string, util.Resources, error) {
	if response, err := self.client.StartDeploymentModification(context.TODO(), &api.StartDeploymentModification{DeploymentId: deploymentId}); err == nil {
		if resources, err := util.DecodeResources(response.ResourcesFormat, response.Resources); err == nil {
			return response.Started, response.NotStartedReason, response.ModificationToken, resources, nil
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
	if response, err := self.client.EndDeploymentModification(context.TODO(), &api.EndDeploymentModification{
		ModificationToken: modificationToken,
		ResourcesFormat:   resourcesFormat,
		Resources:         resources,
	}); err == nil {
		return response.Modified, response.NotModifiedReason, response.DeploymentId, nil
	} else {
		return false, "", "", err
	}
}

func (self *Client) CancelDeploymentModification(modificationToken string) (bool, string, error) {
	if response, err := self.client.CancelDeploymentModification(context.TODO(), &api.CancelDeploymentModification{ModificationToken: modificationToken}); err == nil {
		return response.Cancelled, response.NotCancelledReason, nil
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
