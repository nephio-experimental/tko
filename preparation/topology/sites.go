package topology

import (
	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/preparation"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

var SitesGVK = tkoutil.NewGVK("topology.nephio.org", "v1alpha1", "Sites")

// TODO: cache result
func GetSiteIDs(preparationContext *preparation.Context, name string) ([]string, bool) {
	if site, ok := SitesGVK.NewResourceIdentifier(name).GetResource(preparationContext.DeploymentResources); ok {
		spec := ard.With(site).Get("spec").ConvertSimilar()

		if selectMetadata, ok := spec.Get("select", "metadata").StringMap(); ok {
			metadataPatterns := make(map[string]string)
			for key, value := range selectMetadata {
				metadataPatterns[key] = util.ToString(value)
			}

			if siteInfos, err := preparationContext.Preparation.Client.ListSites(clientpkg.ListSites{MetadataPatterns: metadataPatterns}); err == nil {
				var siteIds []string
				if err := util.IterateResults(siteInfos, func(siteInfo clientpkg.SiteInfo) error {
					siteIds = append(siteIds, siteInfo.SiteID)
					return nil
				}); err != nil {
					preparationContext.Log.Error(err.Error())
				}
				return siteIds, true
			}
		}
	}

	return nil, false
}
