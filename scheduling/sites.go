package scheduling

import (
	contextpkg "context"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

func (self *Scheduling) ScheduleSites() error {
	//self.Log.Notice("scheduling sites")
	if siteInfos, err := self.Client.ListSites(client.ListSites{}); err == nil {
		return util.IterateResults(siteInfos, func(siteInfo client.SiteInfo) error {
			self.ScheduleSite(siteInfo)
			return nil
		})
	} else {
		return err
	}
}

func (self *Scheduling) ScheduleSite(siteInfo client.SiteInfo) {
	log := commonlog.NewKeyValueLogger(self.Log,
		"site", siteInfo.SiteID)
	log.Notice("scheduling site")

	if site, ok, err := self.Client.GetSite(siteInfo.SiteID); err == nil {
		if ok {
			self.scheduleSite(siteInfo.SiteID, site.Package, siteInfo.DeploymentIDs, log)
		} else {
			log.Info("site disappeared")
		}
	} else {
		log.Error(err.Error())
	}
}

func (self *Scheduling) scheduleSite(siteId string, sitePackage tkoutil.Package, deploymentIds []string, log commonlog.Logger) {
	for _, resource := range sitePackage {
		if resourceIdentifier, ok := tkoutil.NewResourceIdentifierForResource(resource); ok {
			if schedulers, err := self.GetSchedulers(resourceIdentifier.GVK); err == nil {
				if len(schedulers) > 0 {
					deployments := make(map[string]tkoutil.Package)
					for _, deploymentId := range deploymentIds {
						if deployment, ok, err := self.Client.GetDeployment(deploymentId); err == nil {
							if ok {
								if deployment.Prepared && deployment.Approved {
									deployments[deploymentId] = deployment.Package
								}
							}
						}
					}

					schedulingContext := self.NewContext(siteId, sitePackage, resourceIdentifier, deployments, log)

					for _, schedule := range schedulers {
						context, cancel := contextpkg.WithTimeout(contextpkg.Background(), self.Timeout)
						if err := schedule(context, schedulingContext); err != nil {
							log.Error(err.Error())
						}
						cancel()
					}
				}
			} else {
				log.Error(err.Error())
			}
		}
	}
}
