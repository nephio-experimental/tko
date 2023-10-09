package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(deploymentCommand)
}

var deploymentCommand = &cobra.Command{
	Use:   "deployment",
	Short: "Work with deployments",
}
