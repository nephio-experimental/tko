package client

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	api "github.com/nephio-experimental/tko/api/grpc"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type DeploymentInfo struct {
	DeploymentID       string            `json:"deploymentId" yaml:"deploymentId"`
	ParentDeploymentID string            `json:"parentDeploymentId,omitempty" yaml:"parentDeploymentId,omitempty"`
	TemplateID         string            `json:"templateId" yaml:"templateId"`
	SiteID             string            `json:"siteId,omitempty" yaml:"siteId,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Prepared           bool              `json:"prepared" yaml:"prepared"`
	Approved           bool              `json:"approved" yaml:"approved"`
}

type Deployment struct {
	DeploymentInfo
	Resources tkoutil.Resources `json:"resources" yaml:"resources"`
}

func (self *Client) CreateDeployment(parentDeploymentId string, templateId string, siteId string, mergeMetadata map[string]string, prepared bool, approved bool, mergeResources tkoutil.Resources) (bool, string, string, error) {
	if mergeResources_, err := self.encodeResources(mergeResources); err == nil {
		return self.CreateDeploymentRaw(parentDeploymentId, templateId, siteId, mergeMetadata, prepared, approved, self.ResourcesFormat, mergeResources_)
	} else {
		return false, "", "", err
	}
}

func (self *Client) CreateDeploymentRaw(parentDeploymentId string, templateId string, siteId string, mergeMetadata map[string]string, prepared bool, approved bool, mergeResourcesFormat string, mergeResources []byte) (bool, string, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Infof("createDeployment: %s, %s, %s, %v, %t, %t, %s", parentDeploymentId, templateId, siteId, mergeMetadata, prepared, approved, mergeResourcesFormat)
		if response, err := apiClient.CreateDeployment(context, &api.CreateDeployment{
			ParentDeploymentId:   parentDeploymentId,
			TemplateId:           templateId,
			SiteId:               siteId,
			MergeMetadata:        mergeMetadata,
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

		self.log.Infof("getDeployment: %s", deploymentId)
		if deployment, err := apiClient.GetDeployment(context, &api.GetDeployment{DeploymentId: deploymentId, PreferredResourcesFormat: self.ResourcesFormat}); err == nil {
			if resources, err := tkoutil.DecodeResources(deployment.ResourcesFormat, deployment.Resources); err == nil {
				return Deployment{
					DeploymentInfo: DeploymentInfo{
						DeploymentID: deployment.DeploymentId,
						TemplateID:   deployment.TemplateId,
						SiteID:       deployment.SiteId,
						Metadata:     deployment.Metadata,
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

		self.log.Infof("deleteDeployment: %s", deploymentId)
		if response, err := apiClient.DeleteDeployment(context, &api.DeploymentID{DeploymentId: deploymentId}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type ListDeployments struct {
	ParentDeploymentID       *string
	TemplateIDPatterns       []string
	TemplateMetadataPatterns map[string]string
	SiteIDPatterns           []string
	SiteMetadataPatterns     map[string]string
	MetadataPatterns         map[string]string
	Prepared                 *bool
	Approved                 *bool
}

// ([fmt.Stringer] interface)
func (self ListDeployments) String() string {
	var s []string
	if self.ParentDeploymentID != nil {
		s = append(s, "parentDeploymentID="+*self.ParentDeploymentID)
	}
	if len(self.TemplateIDPatterns) > 0 {
		s = append(s, "templateIdPatterns="+stringifyStringList(self.TemplateIDPatterns))
	}
	if (self.TemplateMetadataPatterns != nil) && (len(self.TemplateMetadataPatterns) > 0) {
		s = append(s, "templateMetadataPatterns="+stringifyStringMap(self.TemplateMetadataPatterns))
	}
	if len(self.SiteIDPatterns) > 0 {
		s = append(s, "siteIdPatterns="+stringifyStringList(self.TemplateIDPatterns))
	}
	if (self.SiteMetadataPatterns != nil) && (len(self.SiteMetadataPatterns) > 0) {
		s = append(s, "siteMetadataPatterns="+stringifyStringMap(self.SiteMetadataPatterns))
	}
	if (self.MetadataPatterns != nil) && (len(self.MetadataPatterns) > 0) {
		s = append(s, "metadataPatterns="+stringifyStringMap(self.MetadataPatterns))
	}
	if self.Prepared != nil {
		s = append(s, "prepared="+strconv.FormatBool(*self.Prepared))
	}
	if self.Approved != nil {
		s = append(s, "approved="+strconv.FormatBool(*self.Approved))
	}
	return strings.Join(s, " ")
}

func (self *Client) ListDeployments(listDeployments ListDeployments) (util.Results[DeploymentInfo], error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)

		self.log.Infof("listDeployments: %s", listDeployments)
		if client, err := apiClient.ListDeployments(context, &api.ListDeployments{
			ParentDeploymentId:       listDeployments.ParentDeploymentID,
			TemplateIdPatterns:       listDeployments.TemplateIDPatterns,
			TemplateMetadataPatterns: listDeployments.TemplateMetadataPatterns,
			SiteIdPatterns:           listDeployments.SiteIDPatterns,
			SiteMetadataPatterns:     listDeployments.SiteMetadataPatterns,
			MetadataPatterns:         listDeployments.MetadataPatterns,
			Prepared:                 listDeployments.Prepared,
			Approved:                 listDeployments.Approved,
		}); err == nil {
			stream := util.NewResultsStream[DeploymentInfo](cancel)

			go func() {
				for {
					if response, err := client.Recv(); err == nil {
						stream.Send(DeploymentInfo{
							DeploymentID:       response.DeploymentId,
							ParentDeploymentID: response.ParentDeploymentId,
							TemplateID:         response.TemplateId,
							SiteID:             response.SiteId,
							Metadata:           response.Metadata,
							Prepared:           response.Prepared,
							Approved:           response.Approved,
						})
					} else {
						stream.Close(err) // special handling for io.EOF
						return
					}
				}
			}()

			return stream, nil
		} else {
			cancel()
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Client) StartDeploymentModification(deploymentId string) (bool, string, string, tkoutil.Resources, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Infof("startDeploymentModification: %s", deploymentId)
		if response, err := apiClient.StartDeploymentModification(context, &api.StartDeploymentModification{DeploymentId: deploymentId}); err == nil {
			if resources, err := tkoutil.DecodeResources(response.ResourcesFormat, response.Resources); err == nil {
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

func (self *Client) EndDeploymentModification(modificationToken string, resources tkoutil.Resources) (bool, string, string, error) {
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

		self.log.Infof("endDeploymentModification: %s, %s", modificationToken, resourcesFormat)
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

		self.log.Infof("cancelDeploymentModification: %s", modificationToken)
		if response, err := apiClient.CancelDeploymentModification(context, &api.CancelDeploymentModification{ModificationToken: modificationToken}); err == nil {
			return response.Cancelled, response.NotCancelledReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type ModifyDeploymentFunc func(resources tkoutil.Resources) (bool, tkoutil.Resources, error)

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
