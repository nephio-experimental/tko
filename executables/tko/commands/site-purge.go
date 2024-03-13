package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	siteCommand.AddCommand(sitePurgeCommand)

	sitePurgeCommand.Flags().StringArrayVar(&siteIdPatterns, "id", nil, "filter by site ID pattern")
	sitePurgeCommand.Flags().StringArrayVar(&templateIdPatterns, "template-id", nil, "filter by template ID pattern")
	sitePurgeCommand.Flags().StringToStringVarP(&siteMetadata, "metadata", "m", nil, "filter by metadata")
}

var sitePurgeCommand = &cobra.Command{
	Use:   "purge",
	Short: "Purge sites",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		PurgeSites(siteIdPatterns, templateIdPatterns, siteMetadata)
	},
}

func PurgeSites(siteIdPatterns []string, templateIdPatterns []string, siteMetadataPatterns map[string]string) {
	ok, reason, err := NewClient().PurgeSites(client.SelectSites{
		SiteIDPatterns:     siteIdPatterns,
		TemplateIDPatterns: templateIdPatterns,
		MetadataPatterns:   siteMetadataPatterns,
	})
	FailOnGRPCError(err)
	if !ok {
		util.Fail(reason)
	}
}
