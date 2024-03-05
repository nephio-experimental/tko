package client

import (
	contextpkg "context"
	"strings"
	"time"

	api "github.com/nephio-experimental/tko/api/grpc"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/kutil/util"
)

type SiteInfo struct {
	SiteID        string            `json:"siteId" yaml:"siteId"`
	TemplateID    string            `json:"templateId,omitempty" yaml:"templateId,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Updated       time.Time         `json:"updated" yaml:"updated"`
	DeploymentIDs []string          `json:"deploymentIds,omitempty" yaml:"deploymentIds,omitempty"`
}

type Site struct {
	SiteInfo
	Package tkoutil.Package `json:"package" yaml:"package"`
}

func (self *Client) RegisterSite(siteId string, templateId string, metadata map[string]string, package_ tkoutil.Package) (bool, string, error) {
	if package__, err := self.encodePackage(package_); err == nil {
		return self.RegisterSiteRaw(siteId, templateId, metadata, self.PackageFormat, package__)
	} else {
		return false, "", err
	}
}

func (self *Client) RegisterSiteRaw(siteId string, templateId string, metadata map[string]string, packageFormat string, package_ []byte) (bool, string, error) {
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("registerSite",
			"siteId", siteId,
			"templateId", templateId,
			"metadata", metadata,
			"packageFormat", packageFormat)
		if response, err := apiClient.RegisterSite(context, &api.Site{
			SiteId:        siteId,
			TemplateId:    templateId,
			Metadata:      metadata,
			PackageFormat: packageFormat,
			Package:       package_,
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
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Info("getSite",
			"siteId", siteId)
		if site, err := apiClient.GetSite(context, &api.GetSite{SiteId: siteId, PreferredPackageFormat: self.PackageFormat}); err == nil {
			if package_, err := tkoutil.DecodePackage(site.PackageFormat, site.Package); err == nil {
				return Site{
					SiteInfo: SiteInfo{
						SiteID:        site.SiteId,
						TemplateID:    site.TemplateId,
						Metadata:      site.Metadata,
						Updated:       self.toTime(site.Updated),
						DeploymentIDs: site.DeploymentIds,
					},
					Package: package_,
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
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
		defer cancel()

		self.log.Infof("deleteSite",
			"siteId", siteId)
		if response, err := apiClient.DeleteSite(context, &api.SiteID{SiteId: siteId}); err == nil {
			return response.Deleted, response.NotDeletedReason, nil
		} else {
			return false, "", err
		}
	} else {
		return false, "", err
	}
}

type ListSites struct {
	Offset             uint
	MaxCount           uint
	SiteIDPatterns     []string
	TemplateIDPatterns []string
	MetadataPatterns   map[string]string
}

// ([fmt.Stringer] interface)
func (self ListSites) String() string {
	var s []string
	if len(self.SiteIDPatterns) > 0 {
		s = append(s, "siteIdPatterns="+stringifyStringList(self.SiteIDPatterns))
	}
	if len(self.TemplateIDPatterns) > 0 {
		s = append(s, "templateIdPatterns="+stringifyStringList(self.TemplateIDPatterns))
	}
	if (self.MetadataPatterns != nil) && (len(self.MetadataPatterns) > 0) {
		s = append(s, "metadataPatterns="+stringifyStringMap(self.MetadataPatterns))
	}
	return strings.Join(s, " ")
}

func (self *Client) ListSites(listSites ListSites) (util.Results[SiteInfo], error) {
	if apiClient, err := self.APIClient(); err == nil {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)

		self.log.Info("listSites",
			"listSites", listSites)
		if client, err := apiClient.ListSites(context, &api.ListSites{
			Offset:             uint32(listSites.Offset),
			MaxCount:           uint32(listSites.MaxCount),
			SiteIdPatterns:     listSites.SiteIDPatterns,
			TemplateIdPatterns: listSites.TemplateIDPatterns,
			MetadataPatterns:   listSites.MetadataPatterns,
		}); err == nil {
			stream := util.NewResultsStream[SiteInfo](cancel)

			go func() {
				for {
					if listedSite, err := client.Recv(); err == nil {
						stream.Send(SiteInfo{
							SiteID:        listedSite.SiteId,
							TemplateID:    listedSite.TemplateId,
							Metadata:      listedSite.Metadata,
							Updated:       self.toTime(listedSite.Updated),
							DeploymentIDs: listedSite.DeploymentIds,
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
