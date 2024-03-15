package commands

import (
	contextpkg "context"
	"os"
	"time"

	grpcclient "github.com/nephio-experimental/tko/api/grpc-client"
	grpcserver "github.com/nephio-experimental/tko/api/grpc-server"
	httpserver "github.com/nephio-experimental/tko/api/http-server"
	kubernetesserver "github.com/nephio-experimental/tko/api/kubernetes-server"
	backendpkg "github.com/nephio-experimental/tko/backend"
	"github.com/nephio-experimental/tko/backend/memory"
	"github.com/nephio-experimental/tko/backend/spanner"
	"github.com/nephio-experimental/tko/backend/sql"
	"github.com/nephio-experimental/tko/backend/validating"
	tkoutil "github.com/nephio-experimental/tko/util"
	validationpkg "github.com/nephio-experimental/tko/validation"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	cobrautil "github.com/tliron/kutil/cobra"
	"github.com/tliron/kutil/util"
)

const maxModificationDuration = 10 // seconds

var (
	instanceName        string
	instanceDescription string

	backendName           string
	backendConnection     string
	backendConnectTimeout float64
	backendClean          bool

	grpc              bool
	grpcIpStackString string
	grpcIpStack       util.IPStack
	grpcAddress       string
	grpcPort          uint
	grpcFormat        string
	grpcTimeout       float64

	web              bool
	webTimeout       float64
	webIpStackString string
	webIpStack       util.IPStack
	webAddress       string
	webPort          uint
	webTimezone      string

	kubernetes     bool
	kubernetesPort uint

	logIpStackString string
	logIpStack       util.IPStack
	logAddress       string
	logPort          uint

	validatorTimeout float64

	ResetValidationPluginCacheFrequency = 10 * time.Second
	BackendReleaseTimeout               = 10 * time.Second
)

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().StringVar(&instanceName, "name", "Local", "instance name")
	startCommand.Flags().StringVar(&instanceDescription, "description", "", "instance description")
	startCommand.Flags().StringVarP(&backendName, "backend", "b", "memory", "backend implementation")
	startCommand.Flags().StringVar(&backendConnection, "backend-connection", "postgresql://tko:tko@localhost:5432/tko", "backend connection")
	startCommand.Flags().Float64Var(&backendConnectTimeout, "backend-connection-timeout", 30.0, "backend connection timeout in seconds")
	startCommand.Flags().BoolVar(&backendClean, "backend-clean", false, "clean backend data on startup")
	startCommand.Flags().BoolVar(&grpc, "grpc", true, "start gRPC server")
	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "bind IP stack for gRPC server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "bind IP address for gRPC server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "bind TCP port for gRPC server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding KRM over gRPC (\"yaml\" or \"cbor\")")
	startCommand.Flags().Float64Var(&grpcTimeout, "grpc-timeout", 5.0, "gRPC timeout in seconds")
	startCommand.Flags().BoolVar(&web, "web", true, "start web server")
	startCommand.Flags().Float64Var(&webTimeout, "web-timeout", 5.0, "web read/write timeout in seconds")
	startCommand.Flags().StringVar(&webIpStackString, "web-ip-stack", "dual", "bind IP stack for web server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&webAddress, "web-address", "", "bind IP address for web server")
	startCommand.Flags().UintVar(&webPort, "web-port", 50051, "bind TCP port for web server")
	startCommand.Flags().StringVar(&webTimezone, "web-timezone", "", "web server timezone, e.g. \"UTC\" (empty string for local)")
	startCommand.Flags().BoolVar(&kubernetes, "kubernetes", false, "start Kubernetes aggregated API server")
	startCommand.Flags().UintVar(&kubernetesPort, "kubernetes-port", 50052, "bind TCP port for Kubernetes aggregated API server")
	startCommand.Flags().StringVar(&logIpStackString, "log-ip-stack", "dual", "IP stack for log server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&logAddress, "log-address", "", "bind IP address for log server")
	startCommand.Flags().UintVar(&logPort, "log-port", 50055, "bind TCP port for log server")
	startCommand.Flags().Float64Var(&validatorTimeout, "validator-timeout", 30.0, "validator timeout in seconds")

	cobrautil.SetFlagsFromEnvironment("TKO_", startCommand)
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the TKO API Server",
	Run: func(cmd *cobra.Command, args []string) {
		grpcIpStack = util.IPStack(grpcIpStackString)
		util.FailOnError(grpcIpStack.Validate("grpc-ip-stack"))

		webIpStack = util.IPStack(webIpStackString)
		util.FailOnError(grpcIpStack.Validate("web-ip-stack"))

		logIpStack = util.IPStack(logIpStackString)
		util.FailOnError(logIpStack.Validate("log-ip-stack"))

		Serve()
	},
}

func Serve() {
	// Backend
	var backend backendpkg.Backend
	switch backendName {
	case "memory":
		log.Notice("creating memory backend")
		backend = memory.NewMemoryBackend(maxModificationDuration, commonlog.GetLogger("backend.memory"))

	case "postgresql":
		log.Noticef("creating postgresql backend: %s", backendConnection)
		sqlBackend := sql.NewSQLBackend("pgx", backendConnection, "cbor", maxModificationDuration, commonlog.GetLogger("backend.sql"))
		sqlBackend.DropTablesFirst = backendClean
		backend = sqlBackend

	case "spanner":
		backend = spanner.NewSpannerBackend("/span/tmp/"+os.Getenv("USER")+":database-tliron-codelab", commonlog.GetLogger("backend.spanner"))

	default:
		util.Failf("unsupported backend: %s", backendName)
	}

	var webTimezone_ *time.Location
	if webTimezone != "" {
		var err error
		webTimezone_, err = time.LoadLocation(webTimezone)
		util.FailOnError(err)
	}

	// Client
	client := grpcclient.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, tkoutil.SecondsToDuration(grpcTimeout), commonlog.GetLogger("client"))

	// Wrap backend with validation
	validation, err := validationpkg.NewValidation(client, tkoutil.SecondsToDuration(validatorTimeout), commonlog.GetLogger("validation"), logIpStack, logAddress, int(logPort))
	util.FailOnError(err)
	validationTicker := tkoutil.NewTicker(ResetValidationPluginCacheFrequency, validation.ResetPluginCache)
	util.OnExit(validationTicker.Stop)
	backend = validating.NewValidatingBackend(backend, validation)

	util.FailOnError(func() error {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), tkoutil.SecondsToDuration(backendConnectTimeout))
		defer cancel()
		return backend.Connect(context)
	}())
	util.OnExitError(func() error {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), BackendReleaseTimeout)
		defer cancel()
		return backend.Release(context)
	})

	if grpc {
		grpcServer := grpcserver.NewServer(backend, grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("grpc"))
		grpcServer.InstanceName = instanceName
		grpcServer.InstanceDescription = instanceDescription
		util.FailOnError(grpcServer.Start())
		util.OnExit(grpcServer.Stop)
	}

	if web {
		httpServer, err := httpserver.NewServer(backend, tkoutil.SecondsToDuration(webTimeout), webIpStack, webAddress, int(webPort), webTimezone_, commonlog.GetLogger("http"))
		util.FailOnError(err)
		httpServer.InstanceName = instanceName
		httpServer.InstanceDescription = instanceDescription
		util.FailOnError(httpServer.Start())
		util.OnExit(httpServer.Stop)
	}

	if kubernetes {
		kubernetesServer := kubernetesserver.NewServer(backend, int(kubernetesPort), commonlog.GetLogger("kubernetes"))
		util.FailOnError(kubernetesServer.Start())
		util.OnExit(kubernetesServer.Stop)
	}

	// Block forever
	select {}
}
