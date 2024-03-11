package commands

import (
	contextpkg "context"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentModCommand.AddCommand(deploymentModEndCommand)

	deploymentModEndCommand.Flags().StringVarP(&url, "url", "u", "", "URL for package YAML manifests (can be a local directory or file)")
	deploymentModEndCommand.Flags().BoolVarP(&stdin, "stdin", "s", false, "read package YAML manifests from stdin")
}

var deploymentModEndCommand = &cobra.Command{
	Use:   "end [MODIFICATION TOKEN]",
	Short: "End modification of a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), readPackageTimeout)
		util.OnExit(cancel)

		EndDeploymentModification(context, args[0], url, stdin)
	},
}

func EndDeploymentModification(context contextpkg.Context, modificationToken string, url string, stdin bool) {
	package_, err := readPackage(context, url, stdin)
	util.FailOnError(err)

	ok, reason, deploymentId, err := NewClient().EndDeploymentModification(modificationToken, package_)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("modified deployment: %s", deploymentId)
	} else {
		util.Fail(reason)
	}
}
