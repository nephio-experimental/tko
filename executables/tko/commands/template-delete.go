package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateDeleteCommand)
}

var templateDeleteCommand = &cobra.Command{
	Use:   "delete [TEMPLATE ID]",
	Short: "Delete template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DeleteTemplate(args[0])
	},
}

func DeleteTemplate(templateId string) {
	ok, reason, err := NewClient().DeleteTemplate(templateId)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("deleted template: %s", templateId)
	} else {
		util.Fail(reason)
	}
}
