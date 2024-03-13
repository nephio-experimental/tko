package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templatePurgeCommand)

	templatePurgeCommand.Flags().StringArrayVar(&templateIdPatterns, "id", nil, "filter by template ID pattern")
	templatePurgeCommand.Flags().StringToStringVarP(&templateMetadata, "metadata", "m", nil, "filter by metadata")
}

var templatePurgeCommand = &cobra.Command{
	Use:   "purge",
	Short: "Purge templates",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		PurgeTemplates(templateIdPatterns, templateMetadata)
	},
}

func PurgeTemplates(templateIdPatterns []string, templateMetadataPatterns map[string]string) {
	ok, reason, err := NewClient().PurgeTemplates(client.SelectTemplates{
		TemplateIDPatterns: templateIdPatterns,
		MetadataPatterns:   templateMetadataPatterns,
	})
	FailOnGRPCError(err)
	if !ok {
		util.Fail(reason)
	}
}
