package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentModCommand.AddCommand(deploymentModCancelCommand)
}

var deploymentModCancelCommand = &cobra.Command{
	Use:   "cancel [MODIFICATION TOKEN]",
	Short: "Cancel modification of a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		CancelDeploymentModification(args[0])
	},
}

func CancelDeploymentModification(modificationToken string) {
	ok, reason, err := NewClient().CancelDeploymentModification(modificationToken)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("cancelled deployment modification: %s", modificationToken)
	} else {
		util.Fail(reason)
	}
}
