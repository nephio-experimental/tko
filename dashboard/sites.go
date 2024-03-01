package dashboard

import (
	"slices"
	"strconv"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateSites(table *tview.Table) {
	if siteInfos, err := self.client.ListSites(client.ListSites{}); err == nil {
		if siteInfos_, err := util.GatherResults(siteInfos); err == nil {
			slices.SortFunc(siteInfos_, func(a client.SiteInfo, b client.SiteInfo) int {
				return strings.Compare(a.SiteID, b.SiteID)
			})

			table.Clear()

			SetTableHeader(table, "ID", "Template", "Deployments", "Updated")

			for row, siteInfo := range siteInfos_ {
				row++
				table.SetCell(row, 0, tview.NewTableCell(siteInfo.SiteID).SetReference(&SiteDetails{siteInfo.SiteID, self.client}))
				if siteInfo.TemplateID != "" {
					table.SetCell(row, 1, tview.NewTableCell(siteInfo.TemplateID).SetReference(&TemplateDetails{siteInfo.TemplateID, self.client}))
				} else {
					table.SetCellSimple(row, 1, "")
				}
				table.SetCellSimple(row, 2, strconv.Itoa(len(siteInfo.DeploymentIDs)))
				table.SetCellSimple(row, 3, self.timestamp(siteInfo.Updated))
			}
		}
	}
}

//
// SiteDetails
//

type SiteDetails struct {
	siteId string
	client *client.Client
}

// ([Details] interface)
func (self *SiteDetails) GetTitle() string {
	return "Site: " + self.siteId
}

// ([Details] interface)
func (self *SiteDetails) GetText() string {
	if site, ok, err := self.client.GetSite(self.siteId); err == nil {
		if ok {
			if s, err := transcriber.Stringify(ToSliceAny(site.Resources)); err == nil {
				return s
			} else {
				return err.Error()
			}
		} else {
			return ""
		}
	} else {
		return err.Error()
	}
}
