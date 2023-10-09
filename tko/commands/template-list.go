package commands

import (
	"github.com/spf13/cobra"
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
	templateInfos, err := NewClient().ListTemplates(templateIdPatterns, templateMetadataPatterns)
	FailOnGRPCError(err)
	Print(templateInfos)
}
