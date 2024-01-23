package metascheduling

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

func (self *MetaScheduling) ScheduleSites() error {
	//self.Log.Notice("scheduling sites")
	if siteInfos, err := self.Client.ListSites(nil, nil, nil); err == nil {
		for _, siteInfo := range siteInfos {
			self.ScheduleSite(siteInfo)
		}
		return nil
	} else {
		return err
	}
}

func (self *MetaScheduling) ScheduleSite(siteInfo client.SiteInfo) {
	log := commonlog.NewScopeLogger(self.Log, siteInfo.SiteID)
	log.Noticef("scheduling site %s", siteInfo.SiteID)
	if site, ok, err := self.Client.GetSite(siteInfo.SiteID); err == nil {
		if ok {
			self.scheduleSite(siteInfo.SiteID, site.Resources, siteInfo.DeploymentIDs, log)
		} else {
			log.Infof("site disappeared: %s", siteInfo.SiteID)
		}
	} else {
		log.Error(err.Error())
	}
}

func (self *MetaScheduling) scheduleSite(siteId string, siteResources util.Resources, deploymentIds []string, log commonlog.Logger) {
	for _, resource := range siteResources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if scheduler, ok, err := self.GetScheduler(resourceIdentifier.GVK); err == nil {
				if ok {
					deployments := make(map[string]util.Resources)
					for _, deploymentId := range deploymentIds {
						if deployment, ok, err := self.Client.GetDeployment(deploymentId); err == nil {
							if ok {
								if deployment.Prepared {
									deployments[deploymentId] = deployment.Resources
								}
							}
						}
					}

					schedulingContext := self.NewContext(siteId, siteResources, resourceIdentifier, deployments, log)
					if err := scheduler(schedulingContext); err != nil {
						log.Error(err.Error())
					}
				}
			} else {
				log.Error(err.Error())
			}
		}
	}
}
