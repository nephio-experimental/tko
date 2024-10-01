package commands

import (
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	schedulingpkg "github.com/nephio-experimental/tko/scheduling"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	cobrautil "github.com/tliron/kutil/cobra"
	"github.com/tliron/kutil/util"
)

var (
	interval float64

	grpcIpStackString string
	grpcIpStack       util.IPStack
	grpcAddress       string
	grpcPort          uint
	grpcFormat        string
	grpcTimeout       float64

	logIpStackString string
	logIpStack       util.IPStack
	logAddress       string
	logPort          uint

	schedulerTimeout float64

	ResetSchedulingPluginCacheFrequency = 10 * time.Second
)

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().Float64Var(&interval, "interval", 3.0, "polling interval in seconds")
	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "IP stack for TKO Data gRPC (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "IP address for TKO Data gRPC")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "TCP port for TKO Data gRPC")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding KRM for TKO Data gRPC (\"yaml\" or \"cbor\")")
	startCommand.Flags().Float64Var(&grpcTimeout, "grpc-timeout", 10.0, "gRPC timeout in seconds")
	startCommand.Flags().StringVar(&logAddress, "log-address", "", "bind IP address for log server")
	startCommand.Flags().StringVar(&logIpStackString, "log-ip-stack", "dual", "IP stack for log server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().UintVar(&logPort, "log-port", 50055, "bind TCP port for log server")
	startCommand.Flags().Float64Var(&schedulerTimeout, "scheduler-timeout", 300.0, "scheduler timeout in seconds")

	cobrautil.SetFlagsFromEnvironment("TKO_", startCommand)
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the TKO Meta-Scheduler",
	Run: func(cmd *cobra.Command, args []string) {
		grpcIpStack = util.IPStack(grpcIpStackString)
		util.FailOnError(grpcIpStack.Validate("grpc-ip-stack"))

		logIpStack = util.IPStack(logIpStackString)
		util.FailOnError(logIpStack.Validate("log-ip-stack"))

		Start()
	},
}

func Start() {
	// Client
	client := clientpkg.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, tkoutil.SecondsToDuration(grpcTimeout), commonlog.GetLogger("client"))

	// Scheduling
	scheduling := schedulingpkg.NewScheduling(client, tkoutil.SecondsToDuration(schedulerTimeout), commonlog.GetLogger("scheduling"), logIpStack, logAddress, int(logPort))
	schedulingTicker := tkoutil.NewTicker(ResetSchedulingPluginCacheFrequency, scheduling.ResetPluginCache)
	schedulingTicker.Start()
	util.OnExit(schedulingTicker.Stop)

	// Controller
	controller := schedulingpkg.NewController(scheduling, tkoutil.SecondsToDuration(interval), commonlog.GetLogger("controller"))

	controller.Start()
	util.OnExit(controller.Stop)

	// Block forever
	select {}
}
