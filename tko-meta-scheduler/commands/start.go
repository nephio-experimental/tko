package commands

import (
	clientpkg "github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/metascheduling"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

var grpcIpStackString string
var grpcIpStack util.IPStack
var grpcAddress string
var grpcPort uint
var grpcFormat string

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "IP stack for TKO API Server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "address for TKO API Server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "HTTP/2 port for TKO API Server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources for TKO API Server (\"yaml\" or \"cbor\")")
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
	client, err := clientpkg.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("client"))
	util.FailOnError(err)

	// Controller
	controller := metascheduling.NewController(metascheduling.NewInstantiation(client, commonlog.GetLogger("meta-scheduling")), commonlog.GetLogger("controller"))

	controller.Start()
	util.OnExit(controller.Stop)

	// Block forever
	select {}
}
