package commands

import (
	clientpkg "github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/preparation"
	"github.com/nephio-experimental/tko/preparation/topology"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/nephio-experimental/tko/validation"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

var grpcIpStack string
var grpcAddress string
var grpcPort uint
var grpcFormat string

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().StringVar(&grpcIpStack, "grpc-ip-stack", "dual", "IP stack for tko API Server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "address for tko API Server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "HTTP/2 port for tko API Server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources for tko API Server (\"yaml\" or \"cbor\")")
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the controller",
	Run: func(cmd *cobra.Command, args []string) {
		util.FailOnError(tkoutil.ValidateIPStack(grpcIpStack, "grpc-ip-stack"))
		Serve()
	},
}

func Serve() {
	// Client
	client, err := clientpkg.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("client"))
	util.FailOnError(err)

	validation_, err := validation.NewValidation(client, commonlog.GetLogger("validation"))
	util.FailOnError(err)

	// Controller
	controller := preparation.NewController(preparation.NewPreparation(client, validation_, commonlog.GetLogger("preparation")), commonlog.GetLogger("controller"))

	// Topology preparation
	controller.Preparation.RegisterPreparer(topology.PlacementGVK, topology.PreparePlacement)
	controller.Preparation.RegisterPreparer(topology.SiteGVK, topology.PrepareSite)
	controller.Preparation.RegisterPreparer(topology.ToscaGVK, topology.PrepareTOSCA)

	controller.Start()
	util.OnExit(controller.Stop)

	// Block forever
	select {}
}
