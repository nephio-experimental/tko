package client

import (
	"context"
	"io"

	api "github.com/nephio-experimental/tko/grpc"
	"github.com/nephio-experimental/tko/util"
)

type SiteInfo struct {
	SiteID        string            `json:"siteId" yaml:"siteId"`
	TemplateID    string            `json:"templateId,omitempty" yaml:"templateId,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	DeploymentIDs []string          `json:"deploymentIds,omitempty" yaml:"deploymentIds,omitempty"`
}

type Site struct {
	SiteInfo
	Resources util.Resources `json:"resources" yaml:"resources"`
}

func (self *Client) RegisterSite(siteId string, templateId string, metadata map[string]string, resources util.Resources) (bool, string, error) {
	if resources_, err := self.encodeResources(resources); err == nil {
		return self.RegisterSiteRaw(siteId, templateId, metadata, self.ResourcesFormat, resources_)
	} else {
		return false, "", err
	}
}

func (self *Client) RegisterSiteRaw(siteId string, templateId string, metadata map[string]string, resourcesFormat string, resources []byte) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		if response, err := apiClient.RegisterSite(context.TODO(), &api.Site{
			SiteId:          siteId,
			TemplateId:      templateId,
			Metadata:        metadata,
			ResourcesFormat: resourcesFormat,
			Resources:       resources,
		}); err == nil {
			return response.Registered, response.NotRegisteredReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func (self *Client) GetSite(siteId string) (Site, bool, error) {
	if apiClient, err := self.apiClient(); err == nil {
		if site, err := apiClient.GetSite(context.TODO(), &api.GetSite{SiteId: siteId, PreferredResourcesFormat: self.ResourcesFormat}); err == nil {
			if resources, err := util.DecodeResources(site.ResourcesFormat, site.Resources); err == nil {
				return Site{
					SiteInfo: SiteInfo{
						SiteID:        site.SiteId,
						TemplateID:    site.TemplateId,
						Metadata:      site.Metadata,
						DeploymentIDs: site.DeploymentIds,
					},
					Resources: resources,
				}, true, nil
			} else {
				return Site{}, false, err
			}
		} else if IsNotFoundError(err) {
			return Site{}, false, nil
		} else {
			return Site{}, false, err
		}
	} else {
		return Site{}, false, err
	}
}

func (self *Client) DeleteSite(siteId string) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		if response, err := apiClient.DeleteSite(context.TODO(), &api.DeleteSite{SiteId: siteId}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func (self *Client) ListSites(siteIdPatterns []string, templateIdPatterns []string, metadataPatterns map[string]string) ([]SiteInfo, error) {
	if apiClient, err := self.apiClient(); err == nil {
		if client, err := apiClient.ListSites(context.TODO(), &api.ListSites{
			SiteIdPatterns:     siteIdPatterns,
			TemplateIdPatterns: templateIdPatterns,
			MetadataPatterns:   metadataPatterns,
		}); err == nil {
			var siteInfos []SiteInfo
			for {
				if response, err := client.Recv(); err == nil {
					siteInfos = append(siteInfos, SiteInfo{
						SiteID:        response.SiteId,
						TemplateID:    response.TemplateId,
						Metadata:      response.Metadata,
						DeploymentIDs: response.DeploymentIds,
					})
				} else if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}
			return siteInfos, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
