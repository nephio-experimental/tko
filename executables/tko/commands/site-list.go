package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/backend"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	siteCommand.AddCommand(siteListCommand)

	siteListCommand.Flags().UintVar(&offset, "offset", 0, "fetch results starting at this offset")
	siteListCommand.Flags().UintVar(&maxCount, "max-count", backend.DefaultMaxCount, "maximum number of results to fetch")
	siteListCommand.Flags().StringArrayVar(&siteIdPatterns, "id", nil, "filter by site ID pattern")
	siteListCommand.Flags().StringArrayVar(&templateIdPatterns, "template-id", nil, "filter by template ID pattern")
	siteListCommand.Flags().StringToStringVarP(&siteMetadata, "metadata", "m", nil, "filter by metadata")
}

var siteListCommand = &cobra.Command{
	Use:   "list",
	Short: "List sites",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListSites(offset, maxCount, siteIdPatterns, templateIdPatterns, siteMetadata)
	},
}

func ListSites(offset uint, maxCount uint, siteIdPatterns []string, templateIdPatterns []string, siteMetadataPatterns map[string]string) {
	siteIds, err := NewClient().ListSites(client.SelectSites{
		SiteIDPatterns:     siteIdPatterns,
		TemplateIDPatterns: templateIdPatterns,
		MetadataPatterns:   siteMetadataPatterns,
	}, offset, maxCount)
	FailOnGRPCError(err)
	siteIds_, err := util.GatherResults(siteIds)
	util.FailOnError(err)
	Print(siteIds_)
}
