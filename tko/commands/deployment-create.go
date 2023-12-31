package commands

import (
	contextpkg "context"
	"fmt"

	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var siteId string
var prepared bool

func init() {
	deploymentCommand.AddCommand(deploymentCreateCommand)

	deploymentCreateCommand.Flags().StringVarP(&url, "url", "u", "", "URL for YAML content (can be a local directory or file)")
	deploymentCreateCommand.Flags().BoolVarP(&stdin, "stdin", "i", false, "use YAML merge content from stdin")
	deploymentCreateCommand.Flags().StringVar(&parentDeploymentId, "parent", "", "parent deployment ID")
	deploymentCreateCommand.Flags().StringVarP(&siteId, "site", "s", "", "deployment site ID")
	deploymentCreateCommand.Flags().BoolVarP(&prepared, "prepared", "r", false, "mark deployment as prepared")
}

var deploymentCreateCommand = &cobra.Command{
	Use:   "create [TEMPLATE ID]",
	Short: "Create deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CreateDeployment(contextpkg.TODO(), parentDeploymentId, args[0], siteId, prepared, url, stdin)
	},
}

func CreateDeployment(context contextpkg.Context, parentDeploymentId string, templateId string, siteId string, prepared bool, url string, stdin bool) {
	var mergeResources tkoutil.Resources
	if stdin || (url != "") {
		var err error
		mergeResources, err = readResources(context, url, stdin)
		util.FailOnError(err)
	}

	ok, reason, deploymentId, err := NewClient().CreateDeployment(parentDeploymentId, templateId, siteId, prepared, mergeResources)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("created deployment: %s", deploymentId)
		fmt.Println(deploymentId)
	} else {
		util.Fail(reason)
	}
}
