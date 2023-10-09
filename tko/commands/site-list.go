package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	siteCommand.AddCommand(siteListCommand)

	siteListCommand.Flags().StringArrayVar(&siteIdPatterns, "id", nil, "filter by site ID pattern")
	siteListCommand.Flags().StringArrayVar(&templateIdPatterns, "template-id", nil, "filter by template ID pattern")
	siteListCommand.Flags().StringToStringVarP(&siteMetadata, "metadata", "m", nil, "filter by metadata")
}

var siteListCommand = &cobra.Command{
	Use:   "list",
	Short: "List sites",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListSites(siteIdPatterns, templateIdPatterns, siteMetadata)
	},
}

func ListSites(siteIdPatterns []string, templateIdPatterns []string, siteMetadataPatterns map[string]string) {
	siteIds, err := NewClient().ListSites(siteIdPatterns, templateIdPatterns, siteMetadataPatterns)
	FailOnGRPCError(err)
	Print(siteIds)
}
