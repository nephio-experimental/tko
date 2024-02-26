package commands

import (
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
)

func init() {
	nameCommand.AddCommand(nameFromCommand)
}

var nameFromCommand = &cobra.Command{
	Use:   "from [ID]",
	Short: "Convert from Kubernetes name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		FromName(args[0])
	},
}

func FromName(name string) {
	id, err := tkoutil.FromKubernetesName(name)
	FailOnGRPCError(err)
	Print(id)
}
