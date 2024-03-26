package dashboard

import (
	"slices"
	"strconv"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateTemplates(table *tview.Table) error {
	// TODO: paging
	if templateInfoResults, err := self.client.ListTemplates(client.SelectTemplates{}, 0, -1); err == nil {
		SetTableHeader(table, "ID", "Deployments", "Updated")

		var templateIds []string
		util.IterateResults(templateInfoResults, func(templateInfo client.TemplateInfo) error {
			templateIds = append(templateIds, templateInfo.TemplateID)
			row := FindTemplateRow(table, templateInfo.TemplateID)
			self.SetTemplateRow(table, row, &templateInfo)
			return nil
		})

		CleanTableRows(table, func(row int) bool {
			return slices.Contains(templateIds, GetTemplateRow(table, row))
		})

		return nil
	} else {
		return err
	}
}

func (self *Application) SetTemplateRow(table *tview.Table, row int, templateInfo *client.TemplateInfo) {
	table.SetCell(row, 0, tview.NewTableCell(templateInfo.TemplateID).SetReference(&TemplateDetails{templateInfo.TemplateID, self}))
	table.SetCellSimple(row, 1, strconv.Itoa(len(templateInfo.DeploymentIDs)))
	table.SetCellSimple(row, 2, self.timestamp(templateInfo.Updated))
}

func GetTemplateRow(table *tview.Table, row int) string {
	return table.GetCell(row, 0).GetReference().(*TemplateDetails).templateId
}

func FindTemplateRow(table *tview.Table, templateId string) int {
	rowCount := table.GetRowCount()
	for row := 1; row < rowCount; row++ {
		if templateId == GetTemplateRow(table, row) {
			return row
		}
	}
	return rowCount
}

//
// TemplateDetails
//

type TemplateDetails struct {
	templateId  string
	application *Application
}

// ([Details] interface)
func (self *TemplateDetails) GetTitle() string {
	return "Template: " + self.templateId
}

// ([Details] interface)
func (self *TemplateDetails) GetText() (string, error) {
	if template, ok, err := self.application.client.GetTemplate(self.templateId); err == nil {
		if ok {
			return PackageToYAML(template.Package)
		} else {
			return "", nil
		}
	} else {
		return "", err
	}
}
