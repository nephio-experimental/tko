package commands

import (
	"errors"
	"fmt"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	"github.com/nephio-experimental/tko/backend"
	"github.com/nephio-experimental/tko/backend/validating"
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
	var deploymentInfos util.Results[clientpkg.DeploymentInfo]

	client := NewClient()

	// TODO: it would be more efficient to add an ApproveDeployments API to the backend

	if deploymentId != "" {
		deploymentInfos = util.NewResultsSlice([]clientpkg.DeploymentInfo{{DeploymentID: deploymentId}})
	} else {
		var parentDemploymentId_ *string
		if parentDemploymentId != "" {
			parentDemploymentId_ = &parentDemploymentId
		}

		var err error
		deploymentInfos, err = client.ListDeployments(clientpkg.SelectDeployments{
			ParentDeploymentID:       parentDemploymentId_,
			TemplateIDPatterns:       templateIdPatterns,
			TemplateMetadataPatterns: templateMetadataPatterns,
			SiteIDPatterns:           siteIdPatterns,
			SiteMetadataPatterns:     siteMetadataPatterns,
			MetadataPatterns:         metadataPatterns,
			Prepared:                 &trueBool,  // must be prepared
			Approved:                 &falseBool, // avoid approving if already approved
		}, 0, 0)
		FailOnGRPCError(err)
	}

	approver := util.NewParallelExecutor(validating.ParallelBufferSize, func(deploymentId string) error {
		if approved, err := client.ModifyDeployment(deploymentId, func(package_ tkoutil.Package) (bool, tkoutil.Package, error) {
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
				Print(fmt.Sprintf("approved: %s", deploymentId))
			} else {
				Print(fmt.Sprintf("already approved: %s", deploymentId))
			}
			return nil
		} else {
			// Swallow not-found errors
			if backend.IsNotFoundError(err) {
				return nil
			}
			return err
		}
	})

	approver.Start(validating.ParallelWorkers)

	empty := true
	util.FailOnError(util.IterateResults(deploymentInfos, func(deploymentInfo clientpkg.DeploymentInfo) error {
		empty = false
		approver.Queue(deploymentInfo.DeploymentID)
		return nil
	}))

	util.FailOnError(errors.Join(approver.Wait()...))

	if empty {
		Print("no deployments to approve")
	}
}
