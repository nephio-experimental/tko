package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateListCommand)

	templateListCommand.Flags().StringArrayVar(&templateIdPatterns, "id", nil, "filter by template ID pattern")
	templateListCommand.Flags().StringToStringVarP(&templateMetadata, "metadata", "m", nil, "filter by metadata")
}

var templateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List templates",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListTemplates(templateIdPatterns, templateMetadata)
	},
}

func ListTemplates(templateIdPatterns []string, templateMetadataPatterns map[string]string) {
	templateInfos, err := NewClient().ListTemplates(client.ListTemplates{
		TemplateIDPatterns: templateIdPatterns,
		MetadataPatterns:   templateMetadataPatterns,
	})
	FailOnGRPCError(err)
	templateInfos_, err := util.GatherResults(templateInfos)
	util.FailOnError(err)
	Print(templateInfos_)
}
