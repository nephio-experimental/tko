package instantiation

import (
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
)

func (self *Instantiation) InstantiateSites() error {
	//self.Log.Notice("instantiating sites")
	if siteInfos, err := self.Client.ListSites(nil, nil, nil); err == nil {
		for _, siteInfo := range siteInfos {
			self.InstantiateSite(siteInfo)
		}
		return nil
	} else {
		return err
	}
}

func (self *Instantiation) InstantiateSite(siteInfo client.SiteInfo) {
	log := commonlog.NewScopeLogger(self.Log, siteInfo.SiteID)
	log.Noticef("instantiating site %s", siteInfo.SiteID)
	if site, ok, err := self.Client.GetSite(siteInfo.SiteID); err == nil {
		if ok {
			self.instantiateSite(siteInfo.SiteID, site.Resources, siteInfo.DeploymentIDs, log)
		} else {
			log.Infof("site disappeared: %s", siteInfo.SiteID)
		}
	} else {
		log.Error(err.Error())
	}
}

func (self *Instantiation) instantiateSite(siteId string, siteResources []util.Resource, deploymentIds []string, log commonlog.Logger) {
	for _, resource := range siteResources {
		if resourceIdentifier, ok := util.NewResourceIdentifierForResource(resource); ok {
			if instantiator, ok, err := self.GetInstantiator(resourceIdentifier.GVK); err == nil {
				if ok {
					deployments := make(map[string][]util.Resource)
					for _, deploymentId := range deploymentIds {
						if deployment, ok, err := self.Client.GetDeployment(deploymentId); err == nil {
							if ok {
								if deployment.Prepared {
									deployments[deploymentId] = deployment.Resources
								}
							}
						}
					}

					context := self.NewContext(siteId, siteResources, resourceIdentifier, deployments, log)
					if err := instantiator(context); err != nil {
						log.Error(err.Error())
					}
				}
			} else {
				log.Error(err.Error())
			}
		}
	}
}
