package metascheduling

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

func (self *MetaScheduling) ScheduleSites() error {
	//self.Log.Notice("scheduling sites")
	if siteInfos, err := self.Client.ListSites(nil, nil, nil); err == nil {
		for _, siteInfo := range siteInfos {
			context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
			self.ScheduleSite(context, siteInfo)
			cancel()
		}
		return nil
	} else {
		return err
	}
}

func (self *MetaScheduling) ScheduleSite(context contextpkg.Context, siteInfo client.SiteInfo) {
	log := commonlog.NewKeyValueLogger(self.Log,
		"site", siteInfo.SiteID)

	log.Notice("scheduling site")
	if site, ok, err := self.Client.GetSite(siteInfo.SiteID); err == nil {
		if ok {
			self.scheduleSite(context, siteInfo.SiteID, site.Resources, siteInfo.DeploymentIDs, log)
		} else {
			log.Info("site disappeared")
		}
	} else {
		log.Error(err.Error())
	}
}

func (self *MetaScheduling) scheduleSite(context contextpkg.Context, siteId string, siteResources util.Resources, deploymentIds []string, log commonlog.Logger) {
	for _, resource := range siteResources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if schedule, ok, err := self.GetScheduler(resourceIdentifier.GVK); err == nil {
				if ok {
					deployments := make(map[string]util.Resources)
					for _, deploymentId := range deploymentIds {
						if deployment, ok, err := self.Client.GetDeployment(deploymentId); err == nil {
							if ok {
								if deployment.Prepared && deployment.Approved {
									deployments[deploymentId] = deployment.Resources
								}
							}
						}
					}

					schedulingContext := self.NewContext(siteId, siteResources, resourceIdentifier, deployments, log)
					if err := schedule(context, schedulingContext); err != nil {
						log.Error(err.Error())
					}
				}
			} else {
				log.Error(err.Error())
			}
		}
	}
}
