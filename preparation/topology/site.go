package topology

import (
	contextpkg "context"
	"errors"
	"fmt"

	"github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/preparation"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

var SiteGVK = util.NewGVK("topology.nephio.org", "v1alpha1", "Site")

// ([preparation.PrepareFunc] signature)
func PrepareSite(context contextpkg.Context, preparationContext *preparation.Context) (bool, util.Resources, error) {
	preparationContext.Log.Info("preparing topology.nephio.org Site",
		"resource", preparationContext.TargetResourceIdentifer)

	// TODO: check that all Placements have been prepared first?

	if site, ok := preparationContext.GetResource(); ok {
		prepared := false

		spec := ard.With(site).Get("spec")

		// TODO: support implicit matching

		if siteId, ok := spec.Get("siteId").String(); ok {
			SetStatusSiteID(site, siteId)
			prepared = true
		} else if GetSpecProvisionIfNotFound(spec, site) {
			if _, ok := GetStatusSiteID(site); !ok {
				templateId, _ := spec.Get("provisionTemplateId").String()

				merge, _ := spec.Get("merge").List()
				ok, mergeResources, err := preparationContext.GetMergeResources(merge)
				if err != nil {
					return false, nil, err
				}
				if !ok {
					return false, nil, nil
				}

				siteId := "provisioned/" + backend.NewID()
				if ok, reason, err := preparationContext.Preparation.Client.RegisterSite(siteId, templateId, map[string]string{"type": "provisioned"}, mergeResources); err == nil {
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
			if !util.SetPreparedAnnotation(site, true) {
				return false, nil, errors.New("malformed Site resource")
			}
			return true, preparationContext.DeploymentResources, nil
		}
	}

	return false, preparationContext.DeploymentResources, nil
}

func GetSite(resources util.Resources, siteName string) (util.Resource, bool) {
	return SiteGVK.NewResourceIdentifier(siteName).GetResource(resources)
}

func GetSpecProvisionIfNotFound(spec *ard.Node, resource util.Resource) bool {
	if provisionIfNotFound, ok := spec.Get("provisionIfNotFound").Boolean(); ok {
		return provisionIfNotFound
	}
	return false
}

func GetStatusSiteID(resource util.Resource) (string, bool) {
	siteId, ok := ard.With(resource).Get("status", "siteId").String()
	return siteId, ok
}

func SetStatusSiteID(resource util.Resource, siteId string) {
	var status ard.Map
	var ok bool
	if status, ok = ard.With(resource).Get("status").Map(); !ok {
		status = make(ard.Map)
		resource["status"] = status
	}

	status["siteId"] = siteId
}
