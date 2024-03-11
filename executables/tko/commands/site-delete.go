package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	siteCommand.AddCommand(siteDeleteCommand)
}

var siteDeleteCommand = &cobra.Command{
	Use:   "delete [SITE ID]",
	Short: "Delete site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DeleteSite(args[0])
	},
}

func DeleteSite(siteId string) {
	ok, reason, err := NewClient().DeleteSite(siteId)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("deleted site: %s", siteId)
	} else {
		util.Fail(reason)
	}
}
