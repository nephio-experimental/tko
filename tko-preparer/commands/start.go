package commands

import (
	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	preparationpkg "github.com/nephio-experimental/tko/preparation"
	"github.com/nephio-experimental/tko/preparation/topology"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
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
var preparerTimeout float64
var validatorTimeout float64

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().Float64Var(&interval, "interval", 3.0, "polling interval in seconds")
	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "IP stack for TKO API Server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "address for TKO API Server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "HTTP/2 port for TKO API Server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources for TKO API Server (\"yaml\" or \"cbor\")")
	startCommand.Flags().Float64Var(&grpcTimeout, "grpc-timeout", 10.0, "gRPC timeout in seconds")
	startCommand.Flags().Float64Var(&preparerTimeout, "preparer-timeout", 30.0, "preparer timeout in seconds")
	startCommand.Flags().Float64Var(&validatorTimeout, "validator-timeout", 30.0, "validator timeout in seconds")
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the TKO Preparer",
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

	// Validation
	validation, err := validationpkg.NewValidation(client, tkoutil.SecondsToDuration(validatorTimeout), commonlog.GetLogger("validation"))
	util.FailOnError(err)

	// Controller
	preparation := preparationpkg.NewPreparation(client, validation, tkoutil.SecondsToDuration(preparerTimeout), commonlog.GetLogger("preparation"))
	controller := preparationpkg.NewController(preparation, tkoutil.SecondsToDuration(interval), commonlog.GetLogger("controller"))

	// Topology preparation
	controller.Preparation.RegisterPreparer(topology.PlacementGVK, topology.PreparePlacement)
	controller.Preparation.RegisterPreparer(topology.SiteGVK, topology.PrepareSite)
	controller.Preparation.RegisterPreparer(topology.TOSCAGVK, topology.PrepareTOSCA)

	controller.Start()
	util.OnExit(controller.Stop)

	// Block forever
	select {}
}
