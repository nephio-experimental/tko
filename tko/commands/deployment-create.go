package commands

import (
	contextpkg "context"

	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var mergeUrl string
var siteId string
var prepared bool
var approved bool

func init() {
	deploymentCommand.AddCommand(deploymentCreateCommand)

	deploymentCreateCommand.Flags().StringToStringVarP(&deploymentMetadata, "mergeable metadata", "m", nil, "metadata")
	deploymentCreateCommand.Flags().StringVar(&mergeUrl, "merge", "", "URL for mergeable YAML content (can be a local directory or file)")
	deploymentCreateCommand.Flags().BoolVarP(&stdin, "stdin", "i", false, "use mergeable YAML content from stdin")
	deploymentCreateCommand.Flags().StringVar(&parentDeploymentId, "parent", "", "parent deployment ID")
	deploymentCreateCommand.Flags().StringVarP(&siteId, "site", "s", "", "deployment site ID")
	deploymentCreateCommand.Flags().BoolVar(&prepared, "prepared", false, "mark deployment as prepared")
	deploymentCreateCommand.Flags().BoolVar(&approved, "approved", false, "mark deployment as approved")
}

var deploymentCreateCommand = &cobra.Command{
	Use:   "create [TEMPLATE ID]",
	Short: "Create deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CreateDeployment(contextpkg.TODO(), parentDeploymentId, args[0], siteId, deploymentMetadata, prepared, approved, url, stdin)
	},
}

func CreateDeployment(context contextpkg.Context, parentDeploymentId string, templateId string, siteId string, mergeMetadata map[string]string, prepared bool, approved bool, url string, stdin bool) {
	var mergeResources tkoutil.Resources
	if stdin || (mergeUrl != "") {
		var err error
		mergeResources, err = readResources(context, mergeUrl, stdin)
		util.FailOnError(err)
	}

	ok, reason, deploymentId, err := NewClient().CreateDeployment(parentDeploymentId, templateId, siteId, mergeMetadata, prepared, approved, mergeResources)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("created deployment: %s", deploymentId)
		Print(deploymentId)
	} else {
		util.Fail(reason)
	}
}
