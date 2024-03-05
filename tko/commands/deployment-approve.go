package commands

import (
	"errors"
	"fmt"

	client "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentCommand.AddCommand(deploymentApproveCommand)

	deploymentApproveCommand.Flags().StringVar(&parentDeploymentId, "parent", "", "filter by parent deployment ID")
	deploymentApproveCommand.Flags().StringToStringVar(&deploymentMetadata, "metadata", nil, "filter by metadata")
	deploymentApproveCommand.Flags().StringArrayVar(&templateIdPatterns, "template-id", nil, "filter by template ID pattern")
	deploymentApproveCommand.Flags().StringToStringVar(&templateMetadata, "template-metadata", nil, "filter by template metadata")
	deploymentApproveCommand.Flags().StringArrayVar(&siteIdPatterns, "site-id", nil, "filter by site ID pattern")
	deploymentApproveCommand.Flags().StringToStringVar(&siteMetadata, "site-metadata", nil, "filter by site metadata")
}

var deploymentApproveCommand = &cobra.Command{
	Use:   "approve [[DEPLOYMENT ID]]",
	Short: "Approve deployments",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var deploymentId string
		if len(args) == 1 {
			deploymentId = args[0]
		}

		ApproveDeployment(deploymentId, parentDeploymentId, templateIdPatterns, templateMetadata, siteIdPatterns, siteMetadata, deploymentMetadata)
	},
}

func ApproveDeployment(deploymentId string, parentDemploymentId string, templateIdPatterns []string, templateMetadataPatterns map[string]string, siteIdPatterns []string, siteMetadataPatterns map[string]string, metadataPatterns map[string]string) {
	var deploymentInfos util.Results[client.DeploymentInfo]

	if deploymentId != "" {
		deploymentInfos = util.NewResultsSlice([]client.DeploymentInfo{{DeploymentID: deploymentId}})
	} else {
		var parentDemploymentId_ *string
		if parentDemploymentId != "" {
			parentDemploymentId_ = &parentDemploymentId
		}

		var err error
		deploymentInfos, err = NewClient().ListDeployments(client.ListDeployments{
			ParentDeploymentID:       parentDemploymentId_,
			TemplateIDPatterns:       templateIdPatterns,
			TemplateMetadataPatterns: templateMetadataPatterns,
			SiteIDPatterns:           siteIdPatterns,
			SiteMetadataPatterns:     siteMetadataPatterns,
			MetadataPatterns:         metadataPatterns,
			Prepared:                 &trueBool,  // must be prepared
			Approved:                 &falseBool, // avoid approving if already approved
		})
		FailOnGRPCError(err)
	}

	empty := true
	util.FailOnError(util.IterateResults(deploymentInfos, func(deploymentInfo client.DeploymentInfo) error {
		empty = false
		if approved, err := NewClient().ModifyDeployment(deploymentInfo.DeploymentID, func(package_ tkoutil.Package) (bool, tkoutil.Package, error) {
			if deployment, ok := tkoutil.DeploymentResourceIdentifier.GetResource(package_); ok {
				if tkoutil.SetApprovedAnnotation(deployment, true) {
					return true, package_, nil
				} else {
					return false, nil, nil
				}
			} else {
				return false, nil, errors.New("malformed Deployment")
			}
		}); err == nil {
			if approved {
				Print(fmt.Sprintf("approved: %s", deploymentInfo.DeploymentID))
			} else {
				Print(fmt.Sprintf("already approved: %s", deploymentInfo.DeploymentID))
			}
			return nil
		} else {
			return err
		}
	}))

	if empty {
		Print("no deployments to approve")
	}
}
