package commands

import (
	"os"

	backendpkg "github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/api/backend/memory"
	"github.com/nephio-experimental/tko/api/backend/spanner"
	"github.com/nephio-experimental/tko/api/backend/sql"
	"github.com/nephio-experimental/tko/api/client"
	"github.com/nephio-experimental/tko/api/server"
	"github.com/nephio-experimental/tko/validation"
	"github.com/nephio-experimental/tko/web"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

const modificationWindow = 10 // seconds

var backendName string
var grpcProtocol string
var grpcAddress string
var grpcPort uint
var grpcFormat string
var webProtocol string
var webAddress string
var webPort uint

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().StringVarP(&backendName, "backend", "b", "memory", "backend implementation")
	startCommand.Flags().StringVar(&grpcProtocol, "grpc-protocol", "tcp", "protocol for gRPC server")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "address for gRPC server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "HTTP/2 port for gRPC server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources over gRPC (\"yaml\" or \"cbor\")")
	startCommand.Flags().StringVar(&webProtocol, "web-protocol", "tcp", "protocol for web server")
	startCommand.Flags().StringVar(&webAddress, "web-address", "", "address for web server")
	startCommand.Flags().UintVar(&webPort, "web-port", 50051, "HTTP/2 port for web server")
}

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
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
	client_, err := client.NewClient(grpcProtocol, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("client"))
	util.FailOnError(err)

	validation_, err := validation.NewValidation(client_, commonlog.GetLogger("validation"))
	util.FailOnError(err)

	// Wrap backend with validation
	backend = backendpkg.NewValidatingBackend(backend, validation_)
	err = backend.Connect()
	util.FailOnError(err)
	util.OnExitError(backend.Release)

	grpcServer := server.NewServer(backend, grpcProtocol, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("grpc"))
	err = grpcServer.Start()
	util.FailOnError(err)
	util.OnExit(grpcServer.Stop)

	webServer, err := web.NewServer(backend, webProtocol, webAddress, int(webPort), commonlog.GetLogger("web"))
	util.FailOnError(err)
	err = webServer.Start()
	util.FailOnError(err)
	util.OnExitError(webServer.Stop)

	// Block forever
	select {}
}
