package client

import (
	contextpkg "context"
	"strings"

	api "github.com/nephio-experimental/tko/api/grpc"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type TemplateInfo struct {
	TemplateID    string            `json:"templateId" yaml:"templateId"`
	Metadata      map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	DeploymentIDs []string          `json:"deploymentIds,omitempty" yaml:"deploymentIds,omitempty"`
}

type Template struct {
	TemplateInfo
	Resources tkoutil.Resources `json:"resources" yaml:"resources"`
}

func (self *Client) RegisterTemplate(templateId string, metadata map[string]string, resources tkoutil.Resources) (bool, string, error) {
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

		self.log.Infof("registerTemplate: %s, %v, %s", templateId, metadata, resourcesFormat)
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

		self.log.Infof("getTemplate: %s", templateId)
		if template, err := apiClient.GetTemplate(context, &api.GetTemplate{TemplateId: templateId, PreferredResourcesFormat: self.ResourcesFormat}); err == nil {
			if resources, err := tkoutil.DecodeResources(template.ResourcesFormat, template.Resources); err == nil {
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

		self.log.Infof("deleteTemplate: %s", templateId)
		if response, err := apiClient.DeleteTemplate(context, &api.TemplateID{TemplateId: templateId}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type ListTemplates struct {
	TemplateIDPatterns []string
	MetadataPatterns   map[string]string
}

// ([fmt.Stringer] interface)
func (self ListTemplates) String() string {
	var s []string
	if len(self.TemplateIDPatterns) > 0 {
		s = append(s, "templateIdPatterns="+strings.Join(self.TemplateIDPatterns, ","))
	}
	if (self.MetadataPatterns != nil) && (len(self.MetadataPatterns) > 0) {
		s = append(s, "metadataPatterns="+stringifyStringMap(self.MetadataPatterns))
	}
	return strings.Join(s, " ")
}

func (self *Client) ListTemplates(listTemplates ListTemplates) (util.Results[TemplateInfo], error) {
	if apiClient, err := self.apiClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)

		self.log.Infof("listTemplates: %s", listTemplates)
		if client, err := apiClient.ListTemplates(context, &api.ListTemplates{
			TemplateIdPatterns: listTemplates.TemplateIDPatterns,
			MetadataPatterns:   listTemplates.MetadataPatterns,
		}); err == nil {
			stream := util.NewResultsStream[TemplateInfo](cancel)

			go func() {
				for {
					if response, err := client.Recv(); err == nil {
						stream.Send(TemplateInfo{
							TemplateID:    response.TemplateId,
							Metadata:      response.Metadata,
							DeploymentIDs: response.DeploymentIds,
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
