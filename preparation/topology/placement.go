package topology

import (
	"errors"
	"fmt"

	"github.com/nephio-experimental/tko/preparation"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

var PlacementGVK = util.NewGVK("topology.nephio.org", "v1alpha1", "Placement")

type Deployment struct {
	TemplateID     string
	MergeResources util.Resources
	SiteID         string
	Site           util.Resource
}

// ([preparation.PrepareFunc] signature)
func PreparePlacement(preparationContext *preparation.Context) (bool, util.Resources, error) {
	preparationContext.Log.Infof("preparing topology.nephio.org Placement: %s", preparationContext.TargetResourceIdentifer.Name)

	if placement, ok := preparationContext.GetResource(); ok {
		prepared := true
		var deployments []Deployment

		// Gather deployments
		templates, _ := ard.With(placement).Get("spec", "templates").List()
		for _, template := range templates {
			template_ := ard.With(template)
			if templateName, ok := template_.Get("template").String(); ok {
				if templateId, ok := GetTemplateID(preparationContext.DeploymentResources, templateName); ok {
					merge, _ := template_.Get("merge").List()
					_, mergeResources, err := preparationContext.GetMergeResources(merge)
					if err != nil {
						return false, nil, err
					}

					sites, _ := template_.Get("sites").List()
					for _, site := range sites {
						if siteName, ok := site.(string); ok {
							if site_, ok := GetSite(preparationContext.DeploymentResources, siteName); ok {
								if siteId, ok := GetStatusSiteID(site_); ok {
									deployments = append(deployments, Deployment{templateId, mergeResources, siteId, site_})
								} else {
									// Site is not assigned
									prepared = false
								}
							} else {
								return false, nil, fmt.Errorf("site not found: %s", site)
							}
						} else {
							// Selection
							if metadata, ok := ard.With(site).Get("select", "metadata").Map(); ok {
								metadataPatterns := make(map[string]string)
								for key, value := range metadata {
									metadataPatterns[key.(string)] = value.(string)
								}
								if siteInfos, err := preparationContext.Preparation.Client.ListSites(nil, nil, metadataPatterns); err == nil {
									for _, siteInfo := range siteInfos {
										deployments = append(deployments, Deployment{templateId, mergeResources, siteInfo.SiteID, nil})
									}
								}
							}
						}
					}
				} else {
					return false, nil, fmt.Errorf("template not found: %s", templateName)
				}
			}
		}

		if prepared {
			for _, deployment := range deployments {
				if ok, reason, deploymentId, err := preparationContext.Preparation.Client.CreateDeployment(preparationContext.DeploymentID, deployment.TemplateID, deployment.SiteID, false, deployment.MergeResources); err == nil {
					if ok {
						preparationContext.Log.Infof("created deployment %s (%s) for site %s", deploymentId, deployment.TemplateID, deployment.SiteID)
						/*AppendStatusDeploymentID(placement, deploymentId)
						if deployment.Site != nil {
							AppendStatusDeploymentID(deployment.Site, deploymentId)
						}*/
					} else {
						return false, nil, fmt.Errorf("did not create deployment: %s", reason)
					}
				} else {
					return false, nil, err
				}
			}

			if !util.SetPreparedAnnotation(placement, true) {
				return false, nil, errors.New("malformed Placement resource")
			}
		}

		return true, preparationContext.DeploymentResources, nil
	}

	return false, nil, nil
}

func AppendStatusDeploymentID(resource util.Resource, deploymentId string) {
	var status ard.Map
	var ok bool
	if status, ok = ard.With(resource).Get("status").Map(); !ok {
		status = make(ard.Map)
		resource["status"] = status
	}

	deploymentIds, _ := ard.With(status).Get("deploymentIds").List()
	for _, deploymentId_ := range deploymentIds {
		if deploymentId_ == deploymentId {
			return
		}
	}

	deploymentIds = append(deploymentIds, deploymentId)
	status["deploymentIds"] = deploymentIds
}
