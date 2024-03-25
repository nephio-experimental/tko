package dashboard

import (
	"slices"
	"strconv"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateSites(table *tview.Table) {
	// TODO: paging
	if siteInfoResults, err := self.client.ListSites(client.SelectSites{}, 0, -1); err == nil {
		table.Clear()

		SetTableHeader(table, "ID", "Template", "Deployments", "Updated")

		var siteIds []string
		util.IterateResults(siteInfoResults, func(siteInfo client.SiteInfo) error {
			siteIds = append(siteIds, siteInfo.SiteID)
			row := FindSiteRow(table, siteInfo.SiteID)
			self.SetSiteRow(table, row, &siteInfo)
			return nil
		})

		CleanTableRows(table, func(row int) bool {
			return slices.Contains(siteIds, GetSiteRow(table, row))
		})
	}
}

func (self *Application) SetSiteRow(table *tview.Table, row int, siteInfo *client.SiteInfo) {
	table.SetCell(row, 0, tview.NewTableCell(siteInfo.SiteID).SetReference(&SiteDetails{siteInfo.SiteID, self.client}))
	if siteInfo.TemplateID != "" {
		table.SetCell(row, 1, tview.NewTableCell(siteInfo.TemplateID).SetReference(&TemplateDetails{siteInfo.TemplateID, self.client}))
	} else {
		table.SetCellSimple(row, 1, "")
	}
	table.SetCellSimple(row, 2, strconv.Itoa(len(siteInfo.DeploymentIDs)))
	table.SetCellSimple(row, 3, self.timestamp(siteInfo.Updated))
}

func GetSiteRow(table *tview.Table, row int) string {
	return table.GetCell(row, 0).GetReference().(*SiteDetails).siteId
}

func FindSiteRow(table *tview.Table, siteId string) int {
	rowCount := table.GetRowCount()
	for row := 1; row < rowCount; row++ {
		if siteId == GetSiteRow(table, row) {
			return row
		}
	}
	return rowCount
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
			return PackageToYAML(site.Package)
		} else {
			return ""
		}
	} else {
		return err.Error()
	}
}
