package commands

import (
	"time"

	"github.com/nephio-experimental/tko/dashboard"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

var (
	dashboardFrequency float64
	timezone           string
)

func init() {
	rootCommand.AddCommand(dashboardCommand)

	dashboardCommand.Flags().Float64VarP(&dashboardFrequency, "frequency", "f", 3.0, "update frequency in seconds")
	dashboardCommand.Flags().StringVarP(&timezone, "timezone", "t", "", "timezone, e.g. \"UTC\" (empty string for local)")
}

var dashboardCommand = &cobra.Command{
	Use:   "dashboard",
	Short: "Start dashboard TUI",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var timezone_ *time.Location
		if timezone != "" {
			var err error
			timezone_, err = time.LoadLocation(timezone)
			util.FailOnError(err)
		}

		err := dashboard.Dashboard(NewClient(), tkoutil.SecondsToDuration(dashboardFrequency), timezone_)
		util.FailOnError(err)
	},
}
