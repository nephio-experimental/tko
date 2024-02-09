package commands

import (
	"github.com/spf13/cobra"
)

var preparedFilter string
var approvedFilter string

func init() {
	deploymentCommand.AddCommand(deploymentListCommand)

	deploymentListCommand.Flags().StringVar(&parentDeploymentId, "parent", "", "filter by parent deployment ID")
	deploymentListCommand.Flags().StringToStringVar(&deploymentMetadata, "metadata", nil, "filter by metadata")
	deploymentListCommand.Flags().StringArrayVar(&templateIdPatterns, "template-id", nil, "filter by template ID pattern")
	deploymentListCommand.Flags().StringToStringVar(&templateMetadata, "template-metadata", nil, "filter by template metadata")
	deploymentListCommand.Flags().StringArrayVar(&siteIdPatterns, "site-id", nil, "filter by site ID pattern")
	deploymentListCommand.Flags().StringToStringVar(&siteMetadata, "site-metadata", nil, "filter by site metadata")
	deploymentListCommand.Flags().StringVar(&preparedFilter, "prepared", "", "filter by prepared state (\"true\", \"false\", or empty)")
	deploymentListCommand.Flags().StringVar(&approvedFilter, "approved", "", "filter by approved state (\"true\", \"false\", or empty)")
}

var deploymentListCommand = &cobra.Command{
	Use:   "list",
	Short: "List deployments",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ListDeployments(parentDeploymentId, templateIdPatterns, templateMetadata, siteIdPatterns, siteMetadata, deploymentMetadata, preparedFilter, approvedFilter)
	},
}

func ListDeployments(parentDemploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string, metadataPatterns map[string]string, preparedFilter string, approvedFilter string) {
	true_ := true
	false_ := false

	var prepared *bool
	switch preparedFilter {
	case "true":
		prepared = &true_
	case "false":
		prepared = &false_
	}

	var approved *bool
	switch approvedFilter {
	case "true":
		approved = &true_
	case "false":
		approved = &false_
	}

	var parentDemploymentId_ *string
	if parentDemploymentId != "" {
		parentDemploymentId_ = &parentDemploymentId
	}

	deploymentInfos, err := NewClient().ListDeployments(parentDemploymentId_, templateIdPatterns, templateMetadataPatterns, siteIdPatterns, siteMetadataPatterns, metadataPatterns, prepared, approved)
	FailOnGRPCError(err)
	Print(deploymentInfos)
}
