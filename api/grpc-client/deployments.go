package client

import (
	contextpkg "context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	Created            time.Time         `json:"created" yaml:"created"`
	Updated            time.Time         `json:"updated" yaml:"updated"`
	Prepared           bool              `json:"prepared" yaml:"prepared"`
	Approved           bool              `json:"approved" yaml:"approved"`
}

type Deployment struct {
	DeploymentInfo
	Package tkoutil.Package `json:"package" yaml:"package"`
}

func (self *Client) CreateDeployment(parentDeploymentId string, templateId string, siteId string, mergeMetadata map[string]string, prepared bool, approved bool, mergePackage tkoutil.Package) (bool, string, string, error) {
	if mergePackage_, err := self.encodePackage(mergePackage); err == nil {
		return self.CreateDeploymentRaw(parentDeploymentId, templateId, siteId, mergeMetadata, prepared, approved, self.PackageFormat, mergePackage_)
	} else {
		return false, "", "", err
	}
}

func (self *Client) CreateDeploymentRaw(parentDeploymentId string, templateId string, siteId string, mergeMetadata map[string]string, prepared bool, approved bool, mergePackageFormat string, mergePackage []byte) (bool, string, string, error) {
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("createDeployment",
			"parentDeploymentId", parentDeploymentId,
			"templateId", templateId,
			"siteId", siteId,
			"mergeMetadata", mergeMetadata,
			"prepared", prepared,
			"approved", approved,
			"mergePackageFormat", mergePackageFormat)
		if response, err := apiClient.CreateDeployment(context, &api.CreateDeployment{
			ParentDeploymentId: parentDeploymentId,
			TemplateId:         templateId,
			SiteId:             siteId,
			MergeMetadata:      mergeMetadata,
			Prepared:           prepared,
			Approved:           approved,
			MergePackageFormat: mergePackageFormat,
			MergePackage:       mergePackage,
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
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("getDeployment",
			"deploymentId", deploymentId)
		if deployment, err := apiClient.GetDeployment(context, &api.GetDeployment{DeploymentId: deploymentId, PreferredPackageFormat: self.PackageFormat}); err == nil {
			if package_, err := tkoutil.DecodePackage(deployment.PackageFormat, deployment.Package); err == nil {
				return Deployment{
					DeploymentInfo: DeploymentInfo{
						DeploymentID: deployment.DeploymentId,
						TemplateID:   deployment.TemplateId,
						SiteID:       deployment.SiteId,
						Metadata:     deployment.Metadata,
						Created:      self.toTime(deployment.Created),
						Updated:      self.toTime(deployment.Updated),
						Prepared:     deployment.Prepared,
						Approved:     deployment.Approved,
					},
					Package: package_,
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
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("deleteDeployment",
			"deploymentId", deploymentId)
		if response, err := apiClient.DeleteDeployment(context, &api.DeploymentID{DeploymentId: deploymentId}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type SelectDeployments struct {
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
func (self SelectDeployments) String() string {
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

func (self *Client) ListAllDeployments(selectDeployments SelectDeployments) util.Results[DeploymentInfo] {
	return util.CombineResults(func(offset uint) (util.Results[DeploymentInfo], error) {
		return self.ListDeployments(selectDeployments, offset, ChunkSize)
	})
}

func (self *Client) ListDeployments(selectDeployments SelectDeployments, offset uint, maxCount int) (util.Results[DeploymentInfo], error) {
	var window *api.Window
	var err error
	if window, err = newWindow(offset, maxCount); err != nil {
		return nil, err
	}

	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)

		self.log.Info("listDeployments",
			"selectDeployments", selectDeployments)
		if client, err := apiClient.ListDeployments(context, &api.ListDeployments{
			Window: window,
			Select: &api.SelectDeployments{
				ParentDeploymentId:       selectDeployments.ParentDeploymentID,
				TemplateIdPatterns:       selectDeployments.TemplateIDPatterns,
				TemplateMetadataPatterns: selectDeployments.TemplateMetadataPatterns,
				SiteIdPatterns:           selectDeployments.SiteIDPatterns,
				SiteMetadataPatterns:     selectDeployments.SiteMetadataPatterns,
				MetadataPatterns:         selectDeployments.MetadataPatterns,
				Prepared:                 selectDeployments.Prepared,
				Approved:                 selectDeployments.Approved,
			},
		}); err == nil {
			stream := util.NewResultsStream[DeploymentInfo](cancel)

			go func() {
				for {
					if listedDeployment, err := client.Recv(); err == nil {
						stream.Send(DeploymentInfo{
							DeploymentID:       listedDeployment.DeploymentId,
							ParentDeploymentID: listedDeployment.ParentDeploymentId,
							TemplateID:         listedDeployment.TemplateId,
							SiteID:             listedDeployment.SiteId,
							Metadata:           listedDeployment.Metadata,
							Created:            self.toTime(listedDeployment.Created),
							Updated:            self.toTime(listedDeployment.Updated),
							Prepared:           listedDeployment.Prepared,
							Approved:           listedDeployment.Approved,
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

func (self *Client) PurgeDeployments(selectDeployments SelectDeployments) (bool, string, error) {
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("purgeDeployments",
			"selectDeployments", selectDeployments)
		if response, err := apiClient.PurgeDeployments(context, &api.SelectDeployments{
			ParentDeploymentId:       selectDeployments.ParentDeploymentID,
			TemplateIdPatterns:       selectDeployments.TemplateIDPatterns,
			TemplateMetadataPatterns: selectDeployments.TemplateMetadataPatterns,
			SiteIdPatterns:           selectDeployments.SiteIDPatterns,
			SiteMetadataPatterns:     selectDeployments.SiteMetadataPatterns,
			MetadataPatterns:         selectDeployments.MetadataPatterns,
			Prepared:                 selectDeployments.Prepared,
			Approved:                 selectDeployments.Approved,
		}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func (self *Client) StartDeploymentModification(deploymentId string) (bool, string, string, tkoutil.Package, error) {
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("startDeploymentModification",
			"deploymentId", deploymentId)
		if response, err := apiClient.StartDeploymentModification(context, &api.StartDeploymentModification{DeploymentId: deploymentId}); err == nil {
			if package_, err := tkoutil.DecodePackage(response.PackageFormat, response.Package); err == nil {
				return response.Started, response.NotStartedReason, response.ModificationToken, package_, nil
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

func (self *Client) EndDeploymentModification(modificationToken string, package_ tkoutil.Package) (bool, string, string, error) {
	if package__, err := self.encodePackage(package_); err == nil {
		return self.EndDeploymentModificationRaw(modificationToken, self.PackageFormat, package__)
	} else {
		return false, "", "", err
	}
}

func (self *Client) EndDeploymentModificationRaw(modificationToken string, packageFormat string, package_ []byte) (bool, string, string, error) {
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("endDeploymentModification",
			"modificationToken", modificationToken,
			"packageFormat", packageFormat)
		if response, err := apiClient.EndDeploymentModification(context, &api.EndDeploymentModification{
			ModificationToken: modificationToken,
			PackageFormat:     packageFormat,
			Package:           package_,
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
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("cancelDeploymentModification",
			"modificationToken", modificationToken)
		if response, err := apiClient.CancelDeploymentModification(context, &api.CancelDeploymentModification{ModificationToken: modificationToken}); err == nil {
			return response.Cancelled, response.NotCancelledReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type ModifyDeploymentFunc func(package_ tkoutil.Package) (bool, tkoutil.Package, error)

func (self *Client) ModifyDeployment(deploymentId string, modify ModifyDeploymentFunc) (bool, error) {
	if started, reason, modificationToken, package_, err := self.StartDeploymentModification(deploymentId); err == nil {
		if started {
			if modified, package__, err := modify(package_); err == nil {
				if modified {
					if package___, err := self.encodePackage(package__); err == nil {
						if modified, reason, _, err := self.EndDeploymentModificationRaw(modificationToken, self.PackageFormat, package___); err == nil {
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
