package dashboard

import (
	"strconv"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateTemplates(table *tview.Table) {
	// TODO: paging
	if templateInfoResults, err := self.client.ListTemplates(client.SelectTemplates{}, 0, -1); err == nil {
		table.Clear()

		SetTableHeader(table, "ID", "Deployments", "Updated")

		row := 1
		util.IterateResults(templateInfoResults, func(templateInfo client.TemplateInfo) error {
			table.SetCell(row, 0, tview.NewTableCell(templateInfo.TemplateID).SetReference(&TemplateDetails{templateInfo.TemplateID, self.client}))
			table.SetCellSimple(row, 1, strconv.Itoa(len(templateInfo.DeploymentIDs)))
			table.SetCellSimple(row, 2, self.timestamp(templateInfo.Updated))

			row++
			return nil
		})
	}
}

//
// TemplateDetails
//

type TemplateDetails struct {
	templateId string
	client     *client.Client
}

// ([Details] interface)
func (self *TemplateDetails) GetTitle() string {
	return "Template: " + self.templateId
}

// ([Details] interface)
func (self *TemplateDetails) GetText() string {
	if template, ok, err := self.client.GetTemplate(self.templateId); err == nil {
		if ok {
			if s, err := transcriber.Stringify(ToSliceAny(template.Package)); err == nil {
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
