package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(nameCommand)
}

var nameCommand = &cobra.Command{
	Use:   "name",
	Short: "Convert Kubernetes names",
}
