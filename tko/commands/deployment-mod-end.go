package commands

import (
	contextpkg "context"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentModCommand.AddCommand(deploymentModEndCommand)

	deploymentModEndCommand.Flags().StringVarP(&url, "url", "u", "", "URL for YAML content (can be a local directory or file)")
	deploymentModEndCommand.Flags().BoolVarP(&stdin, "stdin", "s", false, "use YAML content from stdin")
}

var deploymentModEndCommand = &cobra.Command{
	Use:   "end [MODIFICATION TOKEN]",
	Short: "End modification of a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		EndDeploymentModification(contextpkg.TODO(), args[0], url, stdin)
	},
}

func EndDeploymentModification(context contextpkg.Context, modificationToken string, url string, stdin bool) {
	resources, err := readResources(context, url, stdin)
	util.FailOnError(err)

	ok, reason, deploymentId, err := NewClient().EndDeploymentModification(modificationToken, resources)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("modified deployment: %s", deploymentId)
	} else {
		util.Fail(reason)
	}
}
