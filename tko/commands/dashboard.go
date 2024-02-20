package commands

import (
	"github.com/nephio-experimental/tko/dashboard"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var dashboardFrequency float64

func init() {
	rootCommand.AddCommand(dashboardCommand)

	dashboardCommand.Flags().Float64VarP(&dashboardFrequency, "frequency", "f", 3.0, "update frequency in seconds")
}

var dashboardCommand = &cobra.Command{
	Use:   "dashboard",
	Short: "Start dashboard TUI",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := dashboard.Dashboard(NewClient(), tkoutil.SecondsToDuration(dashboardFrequency))
		util.FailOnError(err)
	},
}
