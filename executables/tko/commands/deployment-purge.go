package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentCommand.AddCommand(deploymentPurgeCommand)

	deploymentPurgeCommand.Flags().StringVar(&parentDeploymentId, "parent", "", "filter by parent deployment ID")
	deploymentPurgeCommand.Flags().StringToStringVar(&deploymentMetadata, "metadata", nil, "filter by metadata")
	deploymentPurgeCommand.Flags().StringArrayVar(&templateIdPatterns, "template-id", nil, "filter by template ID pattern")
	deploymentPurgeCommand.Flags().StringToStringVar(&templateMetadata, "template-metadata", nil, "filter by template metadata")
	deploymentPurgeCommand.Flags().StringArrayVar(&siteIdPatterns, "site-id", nil, "filter by site ID pattern")
	deploymentPurgeCommand.Flags().StringToStringVar(&siteMetadata, "site-metadata", nil, "filter by site metadata")
	deploymentPurgeCommand.Flags().StringVar(&preparedFilter, "prepared", "", "filter by prepared state (\"true\", \"false\", or empty)")
	deploymentPurgeCommand.Flags().StringVar(&approvedFilter, "approved", "", "filter by approved state (\"true\", \"false\", or empty)")
}

var deploymentPurgeCommand = &cobra.Command{
	Use:   "purge",
	Short: "purge deployments",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		PurgeDeployments(parentDeploymentId, templateIdPatterns, templateMetadata, siteIdPatterns, siteMetadata, deploymentMetadata, preparedFilter, approvedFilter)
	},
}

func PurgeDeployments(parentDemploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string, metadataPatterns map[string]string, preparedFilter string, approvedFilter string) {
	var prepared *bool
	switch preparedFilter {
	case "true":
		prepared = &trueBool
	case "false":
		prepared = &falseBool
	}

	var approved *bool
	switch approvedFilter {
	case "true":
		approved = &trueBool
	case "false":
		approved = &falseBool
	}

	var parentDemploymentId_ *string
	if parentDemploymentId != "" {
		parentDemploymentId_ = &parentDemploymentId
	}

	ok, reason, err := NewClient().PurgeDeployments(client.SelectDeployments{
		ParentDeploymentID:       parentDemploymentId_,
		TemplateIDPatterns:       templateIdPatterns,
		TemplateMetadataPatterns: templateMetadataPatterns,
		SiteIDPatterns:           siteIdPatterns,
		SiteMetadataPatterns:     siteMetadataPatterns,
		MetadataPatterns:         metadataPatterns,
		Prepared:                 prepared,
		Approved:                 approved,
	})
	FailOnGRPCError(err)
	if !ok {
		util.Fail(reason)
	}
}
