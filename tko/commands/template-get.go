package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateGetCommand)
}

var templateGetCommand = &cobra.Command{
	Use:   "get [TEMPLATE ID]",
	Short: "Get template package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetTemplate(args[0])
	},
}

func GetTemplate(templateId string) {
	template, ok, err := NewClient().GetTemplate(templateId)
	FailOnGRPCError(err)
	if ok {
		PrintPackage(template.Package)
	} else {
		util.Fail("not found")
	}
}
