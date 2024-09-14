package commands

import (
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	preparationpkg "github.com/nephio-experimental/tko/preparation"
	"github.com/nephio-experimental/tko/preparation/topology"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	cobrautil "github.com/tliron/kutil/cobra"
	"github.com/tliron/kutil/util"
)

var (
	interval          float64
	grpcIpStackString string
	grpcIpStack       util.IPStack
	grpcAddress       string
	grpcPort          uint
	grpcFormat        string
	grpcTimeout       float64
	logIpStackString  string
	logIpStack        util.IPStack
	logAddress        string
	logPort           uint
	preparerTimeout   float64
	autoApprove       bool

	ResetPreparationPluginCacheFrequency = 10 * time.Second
)

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().Float64Var(&interval, "interval", 3.0, "polling interval in seconds")
	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "IP stack for TKO Data gRPC (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "IP address for TKO Data gRPC")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "TCP port for TKO Data gRPC")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding KRM for TKO Data gRPC (\"yaml\" or \"cbor\")")
	startCommand.Flags().Float64Var(&grpcTimeout, "grpc-timeout", 10.0, "gRPC timeout in seconds")
	startCommand.Flags().StringVar(&logIpStackString, "log-ip-stack", "dual", "IP stack for log server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&logAddress, "log-address", "", "bind IP address for log server")
	startCommand.Flags().UintVar(&logPort, "log-port", 50055, "bind TCP port for log server")
	startCommand.Flags().Float64Var(&preparerTimeout, "preparer-timeout", 30.0, "preparer timeout in seconds")
	startCommand.Flags().BoolVar(&autoApprove, "auto-approve", true, "whether to automatically approve prepared deployments by default")

	cobrautil.SetFlagsFromEnvironment("TKO_", startCommand)
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the TKO Preparer",
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

	// Preparation
	preparation := preparationpkg.NewPreparation(client, tkoutil.SecondsToDuration(preparerTimeout), autoApprove, commonlog.GetLogger("preparation"), logIpStack, logAddress, int(logPort))
	preparationTicker := tkoutil.NewTicker(ResetPreparationPluginCacheFrequency, preparation.ResetPluginCache)
	util.OnExit(preparationTicker.Stop)

	// Controller
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
