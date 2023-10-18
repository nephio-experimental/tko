package topology

import (
	"errors"
	"fmt"

	"github.com/nephio-experimental/tko/preparation"
	"github.com/nephio-experimental/tko/util"
	"github.com/segmentio/ksuid"
	"github.com/tliron/go-ard"
)

var SiteGVK = util.NewGVK("topology.nephio.org", "v1alpha1", "Site")

// ([preparation.PrepareFunc] signature)
func PrepareSite(context *preparation.Context) (bool, util.Resources, error) {
	context.Log.Infof("preparing topology.nephio.org Site: %s", context.TargetResourceIdentifer.Name)

	// TODO: check that all Placements have been prepared first?

	if site, ok := context.GetResource(); ok {
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
				ok, mergeResources, err := context.GetMergeResources(merge)
				if err != nil {
					return false, nil, err
				}
				if !ok {
					return false, nil, nil
				}

				siteId := "provisioned/" + ksuid.New().String()
				if ok, reason, err := context.Preparation.Client.RegisterSite(siteId, templateId, map[string]string{"type": "provisioned"}, mergeResources); err == nil {
					if ok {
						context.Log.Infof("provisioned new site %s for %s", siteId, context.TargetResourceIdentifer.Name)
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
			return true, context.DeploymentResources, nil
		}
	}

	return false, context.DeploymentResources, nil
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
