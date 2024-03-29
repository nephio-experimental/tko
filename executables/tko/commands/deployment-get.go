package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentCommand.AddCommand(deploymentGetCommand)
}

var deploymentGetCommand = &cobra.Command{
	Use:   "get [DEPLOYMENT ID]",
	Short: "Get deployment package",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetDeployment(args[0])
	},
}

func GetDeployment(deploymentId string) {
	deployment, ok, err := NewClient().GetDeployment(deploymentId)
	FailOnGRPCError(err)
	if ok {
		PrintPackage(deployment.Package)
	} else {
		util.Fail("not found")
	}
}
