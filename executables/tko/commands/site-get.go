package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	siteCommand.AddCommand(siteGetCommand)
}

var siteGetCommand = &cobra.Command{
	Use:   "get [SITE ID]",
	Short: "Get site package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetSite(args[0])
	},
}

func GetSite(siteId string) {
	site, ok, err := NewClient().GetSite(siteId)
	FailOnGRPCError(err)
	if ok {
		PrintPackage(site.Package)
	} else {
		util.Fail("not found")
	}
}
