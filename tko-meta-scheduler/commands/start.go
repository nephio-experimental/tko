package commands

import (
	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	metascheduling "github.com/nephio-experimental/tko/meta-scheduling"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

var interval float64
var grpcIpStackString string
var grpcIpStack util.IPStack
var grpcAddress string
var grpcPort uint
var grpcFormat string
var grpcTimeout float64
var schedulerTimeout float64

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().Float64Var(&interval, "interval", 3.0, "polling interval in seconds")
	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "IP stack for TKO API Server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "address for TKO API Server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "HTTP/2 port for TKO API Server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources for TKO API Server (\"yaml\" or \"cbor\")")
	startCommand.Flags().Float64Var(&grpcTimeout, "grpc-timeout", 10.0, "gRPC timeout in seconds")
	startCommand.Flags().Float64Var(&schedulerTimeout, "scheduler-timeout", 30.0, "scheduler timeout in seconds")
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the TKO Meta-Scheduler",
	Run: func(cmd *cobra.Command, args []string) {
		grpcIpStack = util.IPStack(grpcIpStackString)
		util.FailOnError(grpcIpStack.Validate("grpc-ip-stack"))

		Serve()
	},
}

func Serve() {
	// Client
	client, err := clientpkg.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, tkoutil.SecondsToDuration(grpcTimeout), commonlog.GetLogger("client"))
	util.FailOnError(err)

	// Controller
	metaScheduling := metascheduling.NewMetaScheduling(client, tkoutil.SecondsToDuration(schedulerTimeout), commonlog.GetLogger("meta-scheduling"))
	controller := metascheduling.NewController(metaScheduling, tkoutil.SecondsToDuration(interval), commonlog.GetLogger("controller"))

	controller.Start()
	util.OnExit(controller.Stop)

	// Block forever
	select {}
}