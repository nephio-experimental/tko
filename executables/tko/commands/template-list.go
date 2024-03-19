package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/backend"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateListCommand)

	templateListCommand.Flags().UintVar(&offset, "offset", 0, "fetch results starting at this offset")
	templateListCommand.Flags().UintVar(&maxCount, "max-count", backend.DefaultMaxCount, "maximum number of results to fetch")
	templateListCommand.Flags().StringArrayVar(&templateIdPatterns, "id", nil, "filter by template ID pattern")
	templateListCommand.Flags().StringToStringVarP(&templateMetadata, "metadata", "m", nil, "filter by metadata")
}

var templateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List templates",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListTemplates(offset, maxCount, templateIdPatterns, templateMetadata)
	},
}

func ListTemplates(offset uint, maxCount uint, templateIdPatterns []string, templateMetadataPatterns map[string]string) {
	templateInfos, err := NewClient().ListTemplates(client.SelectTemplates{
		TemplateIDPatterns: templateIdPatterns,
		MetadataPatterns:   templateMetadataPatterns,
	}, offset, int(maxCount))
	FailOnGRPCError(err)
	templateInfos_, err := util.GatherResults(templateInfos)
	util.FailOnError(err)
	Print(templateInfos_)
}
