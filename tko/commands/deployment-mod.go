package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	deploymentCommand.AddCommand(deploymentModCommand)
}

var deploymentModCommand = &cobra.Command{
	Use:   "mod",
	Short: "Modify deployments",
}
