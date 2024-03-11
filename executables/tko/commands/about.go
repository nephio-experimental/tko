package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	rootCommand.AddCommand(aboutCommand)
}

var aboutCommand = &cobra.Command{
	Use:   "about",
	Short: "About TKO server",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		About()
	},
}

func About() {
	about, err := NewClient().About()
	util.FailOnError(err)
	Print(about)
}
