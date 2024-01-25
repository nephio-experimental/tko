package client

import (
	contextpkg "context"
	"io"

	api "github.com/nephio-experimental/tko/api/grpc"
	"github.com/nephio-experimental/tko/util"
)

type TemplateInfo struct {
	TemplateID    string            `json:"templateId" yaml:"templateId"`
	Metadata      map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	DeploymentIDs []string          `json:"deploymentIds,omitempty" yaml:"deploymentIds,omitempty"`
}

type Template struct {
	TemplateInfo
	Resources util.Resources `json:"resources" yaml:"resources"`
}

func (self *Client) RegisterTemplate(templateId string, metadata map[string]string, resources util.Resources) (bool, string, error) {
	if resources_, err := self.encodeResources(resources); err == nil {
		return self.RegisterTemplateRaw(templateId, metadata, self.ResourcesFormat, resources_)
	} else {
		return false, "", err
	}
}

func (self *Client) RegisterTemplateRaw(templateId string, metadata map[string]string, resourcesFormat string, resources []byte) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		if response, err := apiClient.RegisterTemplate(context, &api.Template{
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

func (self *Client) GetTemplate(templateId string) (Template, bool, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		if template, err := apiClient.GetTemplate(context, &api.GetTemplate{TemplateId: templateId, PreferredResourcesFormat: self.ResourcesFormat}); err == nil {
			if resources, err := util.DecodeResources(template.ResourcesFormat, template.Resources); err == nil {
				return Template{
					TemplateInfo: TemplateInfo{
						TemplateID:    template.TemplateId,
						Metadata:      template.Metadata,
						DeploymentIDs: template.DeploymentIds,
					},
					Resources: resources,
				}, true, nil
			} else {
				return Template{}, false, err
			}
		} else if IsNotFoundError(err) {
			return Template{}, false, nil
		} else {
			return Template{}, false, err
		}
	} else {
		return Template{}, false, err
	}
}

func (self *Client) DeleteTemplate(templateId string) (bool, string, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		if response, err := apiClient.DeleteTemplate(context, &api.DeleteTemplate{TemplateId: templateId}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

func (self *Client) ListTemplates(templateIdPatterns []string, metadataPatterns map[string]string) ([]TemplateInfo, error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		if client, err := apiClient.ListTemplates(context, &api.ListTemplates{
			TemplateIdPatterns: templateIdPatterns,
			MetadataPatterns:   metadataPatterns,
		}); err == nil {
			var templateInfos []TemplateInfo
			for {
				if response, err := client.Recv(); err == nil {
					templateInfos = append(templateInfos, TemplateInfo{
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
			return templateInfos, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
