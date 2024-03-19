package dashboard

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateDeployments(table *tview.Table) {
	// TODO: paging
	if deploymentnfoResults, err := self.client.ListDeployments(client.SelectDeployments{}, 0, -1); err == nil {
		table.Clear()

		SetTableHeader(table, "ID", "Template", "Parent", "Site", "Prepared", "Approved", "Created", "Updated")

		row := 1
		util.IterateResults(deploymentnfoResults, func(deploymentInfo client.DeploymentInfo) error {
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
			table.SetCell(row, 4, BoolTableCell(deploymentInfo.Prepared))
			table.SetCell(row, 5, BoolTableCell(deploymentInfo.Approved))
			table.SetCellSimple(row, 6, self.timestamp(deploymentInfo.Created))
			table.SetCellSimple(row, 7, self.timestamp(deploymentInfo.Updated))

			row++
			return nil
		})
	}
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
			if s, err := transcriber.Stringify(ToSliceAny(deployment.Package)); err == nil {
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
