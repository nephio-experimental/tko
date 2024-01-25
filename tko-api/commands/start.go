package commands

import (
	"context"
	"os"

	backendpkg "github.com/nephio-experimental/tko/api/backend"
	"github.com/nephio-experimental/tko/api/backend/memory"
	"github.com/nephio-experimental/tko/api/backend/spanner"
	"github.com/nephio-experimental/tko/api/backend/sql"
	grpcclient "github.com/nephio-experimental/tko/api/grpc-client"
	grpcserver "github.com/nephio-experimental/tko/api/grpc-server"
	httpserver "github.com/nephio-experimental/tko/api/http-server"
	"github.com/nephio-experimental/tko/validation"
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/util"
)

const maxModificationDuration = 10 // seconds

var backendName string
var backendClean bool
var grpcIpStackString string
var grpcIpStack util.IPStack
var grpcAddress string
var grpcPort uint
var grpcFormat string
var grpcTimeout float64
var webIpStackString string
var webIpStack util.IPStack
var webAddress string
var webPort uint

func init() {
	rootCommand.AddCommand(startCommand)

	startCommand.Flags().StringVarP(&backendName, "backend", "b", "memory", "backend implementation")
	startCommand.Flags().BoolVarP(&backendClean, "backend-clean", "", false, "clean backend data on startup")
	startCommand.Flags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "bind IP stack for gRPC server (\"dual\", \"ipv6\", or \"ipv4\")")
	startCommand.Flags().StringVar(&grpcAddress, "grpc-address", "", "bind address for gRPC server")
	startCommand.Flags().UintVar(&grpcPort, "grpc-port", 50050, "bind HTTP/2 port for gRPC server")
	startCommand.Flags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources over gRPC (\"yaml\" or \"cbor\")")
	startCommand.Flags().Float64Var(&grpcTimeout, "grpc-timeout", 5.0, "gRPC timeout in seconds")
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
		backend = memory.NewMemoryBackend(maxModificationDuration, commonlog.GetLogger("backend.memory"))

	case "postgresql":
		dataSource := "postgresql://tko:tko@localhost:5432/tko"
		log.Noticef("creating postgresql backend: %s", dataSource)
		sqlBackend := sql.NewSQLBackend("pgx", dataSource, "cbor", maxModificationDuration, commonlog.GetLogger("backend.sql"))
		sqlBackend.DropTablesFirst = backendClean
		backend = sqlBackend

	case "spanner":
		backend = spanner.NewSpannerBackend("/span/tmp/"+os.Getenv("USER")+":database-tliron-codelab", commonlog.GetLogger("backend.spanner"))

	default:
		util.Failf("unsupported backend: %s", backendName)
	}

	// Client
	client, err := grpcclient.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, grpcTimeout, commonlog.GetLogger("client"))
	util.FailOnError(err)

	// Wrap backend with validation
	validation_, err := validation.NewValidation(client, commonlog.GetLogger("validation"))
	util.FailOnError(err)
	backend = backendpkg.NewValidatingBackend(backend, validation_)

	util.FailOnError(backend.Connect(context.TODO()))
	util.OnExitError(func() error {
		return backend.Release(context.TODO())
	})

	// gRPC
	grpcServer := grpcserver.NewServer(backend, grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("grpc"))
	util.FailOnError(grpcServer.Start())
	util.OnExit(grpcServer.Stop)

	// HTTP
	httpServer, err := httpserver.NewServer(backend, webIpStack, webAddress, int(webPort), commonlog.GetLogger("http"))
	util.FailOnError(err)
	util.FailOnError(httpServer.Start())
	util.OnExit(httpServer.Stop)

	// Block forever
	select {}
}
