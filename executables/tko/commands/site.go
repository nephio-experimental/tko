package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(siteCommand)
}

var siteCommand = &cobra.Command{
	Use:   "site",
	Short: "Work with sites",
}
