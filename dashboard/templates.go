package dashboard

import (
	"slices"
	"strconv"
	"strings"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/rivo/tview"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
)

// ([UpdateTableFunc] signature)
func (self *Application) UpdateTemplates(table *tview.Table) {
	if templateInfos, err := self.client.ListTemplates(client.ListTemplates{}); err == nil {
		if templateInfos_, err := util.GatherResults(templateInfos); err == nil {
			slices.SortFunc(templateInfos_, func(a client.TemplateInfo, b client.TemplateInfo) int {
				return strings.Compare(a.TemplateID, b.TemplateID)
			})

			table.Clear()

			SetTableHeader(table, "ID", "Deployments")

			for row, templateInfo := range templateInfos_ {
				row++
				table.SetCell(row, 0, tview.NewTableCell(templateInfo.TemplateID).SetReference(&TemplateDetails{templateInfo.TemplateID, self.client}))
				table.SetCellSimple(row, 1, strconv.Itoa(len(templateInfo.DeploymentIDs)))
			}
		}
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
			if s, err := transcribe.NewTranscriber().Stringify(ToSliceAny(template.Resources)); err == nil {
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
