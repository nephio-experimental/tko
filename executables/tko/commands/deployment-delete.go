package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentCommand.AddCommand(deploymentDeleteCommand)
}

var deploymentDeleteCommand = &cobra.Command{
	Use:   "delete [DEPLOYMENT ID]",
	Short: "Delete deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DeleteDeployment(args[0])
	},
}

func DeleteDeployment(deploymentId string) {
	ok, reason, err := NewClient().DeleteDeployment(deploymentId)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("deleted deployment: %s", deploymentId)
	} else {
		util.Fail(reason)
	}
}
