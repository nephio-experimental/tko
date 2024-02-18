package commands

import (
	client "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
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

var trueBool = true
var falseBool = false

func ListDeployments(parentDemploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string, metadataPatterns map[string]string, preparedFilter string, approvedFilter string) {
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

	deploymentInfos, err := NewClient().ListDeployments(client.ListDeployments{
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
	deploymentInfos_, err := util.GatherResults(deploymentInfos)
	util.FailOnError(err)
	Print(deploymentInfos_)
}
