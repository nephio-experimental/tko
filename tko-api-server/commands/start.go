package commands

import (
	"os"

	backendpkg "github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/api/backend/memory"
	"github.com/nephio-experimental/tko/api/backend/spanner"
	"github.com/nephio-experimental/tko/api/backend/sql"
	clientpkg "github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/api/server"
	"github.com/nephio-experimental/tko/validation"
	"github.com/nephio-experimental/tko/web"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

const modificationWindow = 10 // seconds

var backendName string
var grpcIpStackString string
var grpcIpStack util.IPStack
var grpcAddress string
var grpcPort uint
var grpcFormat string
var webIpStackString string
var webIpStack util.IPStack
var webAddress string
var webPort uint

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().StringVarP(&backendName, "backend", "b", "memory", "backend implementation")
	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "bind IP stack for gRPC server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "bind address for gRPC server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "bind HTTP/2 port for gRPC server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources over gRPC (\"yaml\" or \"cbor\")")
	startCommand.Flags().StringVar(&webIpStackString, "web-ip-stack", "dual", "bind IP stack for web server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&webAddress, "web-address", "", "bind address for web server")
	startCommand.Flags().UintVar(&webPort, "web-port", 50051, "bind HTTP/2 port for web server")
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the TKO API Server",
	Run: func(cmd *cobra.Command, args []string) {
		grpcIpStack = util.IPStack(grpcIpStackString)
		util.FailOnError(grpcIpStack.Validate("grpc-ip-stack"))

		webIpStack = util.IPStack(webIpStackString)
		util.FailOnError(grpcIpStack.Validate("web-ip-stack"))

		Serve()
	},
}

func Serve() {
	// Backend
	var backend backendpkg.Backend
	switch backendName {
	case "memory":
		log.Notice("creating memory backend")
		backend = memory.NewMemoryBackend(modificationWindow, commonlog.GetLogger("backend"))

	case "postgresql":
		dataSource := "postgresql://tko:tko@localhost:5432/tko"
		log.Noticef("creating postgresql backend: %s", dataSource)
		backend = sql.NewSqlBackend("pgx", dataSource, "cbor", modificationWindow, commonlog.GetLogger("sql"))

	case "spanner":
		backend = spanner.NewSpannerBackend("/span/tmp/"+os.Getenv("USER")+":database-tliron-codelab", commonlog.GetLogger("spanner"))

	default:
		util.Failf("unsupported backend: %s", backendName)
	}

	// Client
	client, err := clientpkg.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("client"))
	util.FailOnError(err)

	// Wrap backend with validation
	validation_, err := validation.NewValidation(client, commonlog.GetLogger("validation"))
	util.FailOnError(err)
	backend = backendpkg.NewValidatingBackend(backend, validation_)

	util.FailOnError(backend.Connect())
	util.OnExitError(backend.Release)

	// gRPC
	grpcServer := server.NewServer(backend, grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("grpc"))
	util.FailOnError(grpcServer.Start())
	util.OnExit(grpcServer.Stop)

	// Web
	webServer, err := web.NewServer(backend, webIpStack, webAddress, int(webPort), commonlog.GetLogger("web"))
	util.FailOnError(err)
	util.FailOnError(webServer.Start())
	util.OnExit(webServer.Stop)

	// Block forever
	select {}
}
