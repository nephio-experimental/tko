package dashboard

import (
	"slices"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateDeployments(table *tview.Table) {
	if deploymentInfos, err := self.client.ListDeployments(client.ListDeployments{}); err == nil {
		if deploymentInfos_, err := util.GatherResults(deploymentInfos); err == nil {
			slices.SortFunc(deploymentInfos_, func(a client.DeploymentInfo, b client.DeploymentInfo) int {
				return strings.Compare(a.DeploymentID, b.DeploymentID)
			})

			table.Clear()

			SetTableHeader(table, "ID", "Template", "Parent", "Site", "Prepared", "Approved", "Created", "Updated")

			for row, deploymentInfo := range deploymentInfos_ {
				row++
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
			}
		}
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
			if s, err := transcriber.Stringify(ToSliceAny(deployment.Resources)); err == nil {
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
