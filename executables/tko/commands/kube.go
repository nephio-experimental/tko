package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(kubeCommand)
}

var kubeCommand = &cobra.Command{
	Use:   "kube",
	Short: "Convert Kubernetes names",
}
