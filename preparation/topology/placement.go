package topology

import (
	contextpkg "context"
	"errors"
	"fmt"

	"github.com/nephio-experimental/tko/preparation"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
)

var PlacementGVK = tkoutil.NewGVK("topology.nephio.org", "v1alpha1", "Placement")

type Deployment struct {
	TemplateID     string
	MergeResources tkoutil.Resources
	SiteID         string
	Site           tkoutil.Resource
}

// ([preparation.PrepareFunc] signature)
func PreparePlacement(context contextpkg.Context, preparationContext *preparation.Context) (bool, tkoutil.Resources, error) {
	if placement, ok := preparationContext.GetResource(); ok {
		prepared := true
		var deployments []Deployment

		// Gather deployments
		templates, _ := ard.With(placement).Get("spec", "templates").List()
		for _, template := range templates {
			template_ := ard.With(template).ConvertSimilar()
			if templateName, ok := template_.Get("template").String(); ok {
				if templateId, ok := GetTemplateID(preparationContext, templateName); ok {
					merge, _ := template_.Get("merge").List()
					_, mergeResources, err := preparationContext.GetMergeResources(merge)
					if err != nil {
						return false, nil, err
					}

					siteNames, _ := template_.Get("sites").StringList()
					for _, siteName := range siteNames {
						if siteIds, ok := GetSiteIDs(preparationContext, siteName); ok {
							for _, siteId := range siteIds {
								deployments = append(deployments, Deployment{templateId, mergeResources, siteId, nil})
							}
							continue
						}

						if siteId, ok := GetSiteID(preparationContext, siteName); ok {
							deployments = append(deployments, Deployment{templateId, mergeResources, siteId, nil})
							continue
						}

						return false, nil, fmt.Errorf("site not found: %s", siteName)

						/*if siteName, ok := site.(string); ok {
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
								if siteInfos, err := preparationContext.Preparation.Client.ListSites(client.ListSites{MetadataPatterns: metadataPatterns}); err == nil {
									if err := util.IterateResults(siteInfos, func(siteInfo client.SiteInfo) error {
										deployments = append(deployments, Deployment{templateId, mergeResources, siteInfo.SiteID, nil})
										return nil
									}); err != nil {
										return false, nil, err
									}
								} else {
									return false, nil, err
								}
							}
						}*/
					}
				} else {
					return false, nil, fmt.Errorf("template not found: %s", templateName)
				}
			} else {
				return false, nil, errors.New("template not specified")
			}
		}

		if prepared {
			for _, deployment := range deployments {
				if ok, reason, deploymentId, err := preparationContext.Preparation.Client.CreateDeployment(preparationContext.DeploymentID, deployment.TemplateID, deployment.SiteID, nil, false, false, deployment.MergeResources); err == nil {
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

			if !tkoutil.SetPreparedAnnotation(placement, true) {
				return false, nil, errors.New("malformed Placement resource")
			}
		}

		return true, preparationContext.DeploymentResources, nil
	}

	return false, nil, nil
}

func AppendStatusDeploymentID(resource tkoutil.Resource, deploymentId string) {
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
