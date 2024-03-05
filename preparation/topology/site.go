package topology

import (
	contextpkg "context"
	"errors"
	"fmt"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/backend"
	"github.com/nephio-experimental/tko/preparation"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

var SiteGVK = tkoutil.NewGVK("topology.nephio.org", "v1alpha1", "Site")

// TODO: cache result
func GetSiteID(preparationContext *preparation.Context, name string) (string, bool) {
	if site, ok := SiteGVK.NewResourceIdentifier(name).GetResource(preparationContext.DeploymentPackage); ok {
		return GetStatusSiteID(site)
	}

	return "", false
}

// ([preparation.PrepareFunc] signature)
func PrepareSite(context contextpkg.Context, preparationContext *preparation.Context) (bool, tkoutil.Package, error) {
	if site, ok := preparationContext.GetTargetResource(); ok {
		prepared := false

		spec := ard.With(site).Get("spec").ConvertSimilar()

		if siteId, ok := spec.Get("siteId").String(); ok {
			SetStatusSiteID(site, siteId)
			prepared = true
		} else if selectMetadata, ok := spec.Get("select", "metadata").StringMap(); ok {
			metadataPatterns := make(map[string]string)
			for key, value := range selectMetadata {
				metadataPatterns[key] = util.ToString(value)
			}

			if siteInfos, err := preparationContext.Preparation.Client.ListSites(clientpkg.ListSites{MetadataPatterns: metadataPatterns}); err == nil {
				// First one we find
				if siteInfo, err := siteInfos.Next(); err == nil {
					siteInfos.Release()
					SetStatusSiteID(site, siteInfo.SiteID)
					prepared = true
				} else {
					siteInfos.Release()
					return false, nil, err
				}
			}
		} else if GetSpecProvisionIfNotFound(spec, site) {
			if _, ok := GetStatusSiteID(site); !ok {
				templateId, _ := spec.Get("provisionTemplateId").String()

				merge, _ := spec.Get("merge").List()
				ok, mergePackage, err := preparationContext.GetMergePackage(merge)
				if err != nil {
					return false, nil, err
				}
				if !ok {
					return false, nil, nil
				}

				siteId := "provisioned/" + backend.NewID()
				if ok, reason, err := preparationContext.Preparation.Client.RegisterSite(siteId, templateId, map[string]string{"type": "provisioned"}, mergePackage); err == nil {
					if ok {
						preparationContext.Log.Infof("provisioned new site %s for %s", siteId, preparationContext.TargetResourceIdentifer.Name)
						SetStatusSiteID(site, siteId)
						prepared = true
					} else {
						return false, nil, fmt.Errorf("did not provision new site: %s", reason)
					}
				} else {
					return false, nil, err
				}
			}
		}

		if prepared {
			if !tkoutil.SetPreparedAnnotation(site, true) {
				return false, nil, errors.New("malformed Site resource")
			}
			return true, preparationContext.DeploymentPackage, nil
		}
	}

	return false, preparationContext.DeploymentPackage, nil
}

func GetSite(package_ tkoutil.Package, siteName string) (tkoutil.Resource, bool) {
	return SiteGVK.NewResourceIdentifier(siteName).GetResource(package_)
}

func GetSpecProvisionIfNotFound(spec *ard.Node, resource tkoutil.Resource) bool {
	if provisionIfNotFound, ok := spec.Get("provisionIfNotFound").Boolean(); ok {
		return provisionIfNotFound
	}
	return false
}

func GetStatusSiteID(resource tkoutil.Resource) (string, bool) {
	siteId, ok := ard.With(resource).Get("status", "siteId").String()
	return siteId, ok
}

func SetStatusSiteID(resource tkoutil.Resource, siteId string) {
	var status ard.Map
	var ok bool
	if status, ok = ard.With(resource).Get("status").Map(); !ok {
		status = make(ard.Map)
		resource["status"] = status
	}

	status["siteId"] = siteId
}
