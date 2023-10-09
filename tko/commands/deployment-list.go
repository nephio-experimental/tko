package commands

import (
	"github.com/spf13/cobra"
)

var preparedFilter string

func init() {
	deploymentCommand.AddCommand(deploymentListCommand)

	deploymentListCommand.Flags().StringVar(&preparedFilter, "prepared", "", "filter by prepared state (\"true\", \"false\", or empty)")
	deploymentListCommand.Flags().StringVar(&parentDeploymentId, "parent", "", "filter by parent deployment ID")
	deploymentListCommand.Flags().StringArrayVar(&templateIdPatterns, "template-id", nil, "filter by template ID pattern")
	deploymentListCommand.Flags().StringToStringVar(&templateMetadata, "template-metadata", nil, "filter by template metadata")
	deploymentListCommand.Flags().StringArrayVar(&siteIdPatterns, "site-id", nil, "filter by site ID pattern")
	deploymentListCommand.Flags().StringToStringVar(&siteMetadata, "site-metadata", nil, "filter by site metadata")
}

var deploymentListCommand = &cobra.Command{
	Use:   "list",
	Short: "List deployments",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListDeployments(preparedFilter, parentDeploymentId, templateIdPatterns, templateMetadata, siteIdPatterns, siteMetadata)
	},
}

func ListDeployments(prepared string, parentDemploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string) {
	deploymentInfos, err := NewClient().ListDeployments(prepared, parentDemploymentId, templateIdPatterns, templateMetadataPatterns, siteIdPatterns, siteMetadataPatterns)
	FailOnGRPCError(err)
	Print(deploymentInfos)
}
