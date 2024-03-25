package dashboard

import (
	"slices"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateDeployments(table *tview.Table) {
	// TODO: paging
	if deploymentnfoResults, err := self.client.ListDeployments(client.SelectDeployments{}, 0, -1); err == nil {
		SetTableHeader(table, "ID", "Template", "Parent", "Site", "Prepared", "Approved", "Created", "Updated")

		var deploymentIds []string
		util.IterateResults(deploymentnfoResults, func(deploymentInfo client.DeploymentInfo) error {
			deploymentIds = append(deploymentIds, deploymentInfo.DeploymentID)
			row := FindDeploymentRow(table, deploymentInfo.DeploymentID)
			self.SetDeploymentRow(table, row, &deploymentInfo)
			return nil
		})

		CleanTableRows(table, func(row int) bool {
			return slices.Contains(deploymentIds, GetDeploymentRow(table, row))
		})
	}
}

func (self *Application) SetDeploymentRow(table *tview.Table, row int, deploymentInfo *client.DeploymentInfo) {
	table.SetCell(row, 0, tview.NewTableCell(deploymentInfo.DeploymentID).SetReference(&DeploymentDetails{deploymentInfo.DeploymentID, self.client}))
	if deploymentInfo.TemplateID != "" {
		table.SetCell(row, 1, tview.NewTableCell(deploymentInfo.TemplateID).SetReference(&TemplateDetails{deploymentInfo.TemplateID, self.client}))
	} else {
		table.SetCellSimple(row, 1, "")
	}
	if deploymentInfo.ParentDeploymentID != "" {
		table.SetCell(row, 2, tview.NewTableCell(deploymentInfo.ParentDeploymentID).SetReference(&DeploymentDetails{deploymentInfo.ParentDeploymentID, self.client}))
	} else {
		table.SetCellSimple(row, 2, "")
	}
	if deploymentInfo.SiteID != "" {
		table.SetCell(row, 3, tview.NewTableCell(deploymentInfo.SiteID).SetReference(&SiteDetails{deploymentInfo.SiteID, self.client}))
	} else {
		table.SetCellSimple(row, 3, "")
	}
	table.SetCell(row, 4, NewBoolTableCell(deploymentInfo.Prepared))
	table.SetCell(row, 5, NewBoolTableCell(deploymentInfo.Approved))
	table.SetCellSimple(row, 6, self.timestamp(deploymentInfo.Created))
	table.SetCellSimple(row, 7, self.timestamp(deploymentInfo.Updated))
}

func GetDeploymentRow(table *tview.Table, row int) string {
	return table.GetCell(row, 0).GetReference().(*DeploymentDetails).deploymentId
}

func FindDeploymentRow(table *tview.Table, deploymentId string) int {
	rowCount := table.GetRowCount()
	for row := 1; row < rowCount; row++ {
		if deploymentId == GetDeploymentRow(table, row) {
			return row
		}
	}
	return rowCount
}

//
// DeploymentDetails
//

type DeploymentDetails struct {
	deploymentId string
	client       *client.Client
}

// ([Details] interface)
func (self *DeploymentDetails) GetTitle() string {
	return "Deployment: " + self.deploymentId
}

// ([Details] interface)
func (self *DeploymentDetails) GetText() string {
	if deployment, ok, err := self.client.GetDeployment(self.deploymentId); err == nil {
		if ok {
			return PackageToYAML(deployment.Package)
		} else {
			return ""
		}
	} else {
		return err.Error()
	}
}
