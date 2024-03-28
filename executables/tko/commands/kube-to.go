package commands

import (
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
)

func init() {
	kubeCommand.AddCommand(kubeToCommand)
}

var kubeToCommand = &cobra.Command{
	Use:   "to [ID]",
	Short: "Convert to Kubernetes name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ToName(args[0])
	},
}

func ToName(id string) {
	name, err := tkoutil.ToKubernetesName(id)
	FailOnGRPCError(err)
	Print(name)
}
