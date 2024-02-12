package commands

import (
	"io"
	"os"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const toolName = "tko"

var log = commonlog.GetLogger(toolName)
var clientLog = commonlog.NewScopeLogger(log, "client")

var url string
var stdin bool
var templateIdPatterns []string
var siteIdPatterns []string
var templateMetadata map[string]string
var siteMetadata map[string]string
var deploymentMetadata map[string]string
var parentDeploymentId string

func NewClient() *clientpkg.Client {
	client, err := clientpkg.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, tkoutil.SecondsToDuration(grpcTimeout), clientLog)
	util.FailOnError(err)
	return client
}

func FailOnGRPCError(err error) {
	if status_, ok := status.FromError(err); ok {
		switch code := status_.Code(); code {
		case codes.OK:
			return
		case codes.Unknown:
			util.Fail(status_.Message())
		default:
			util.Failf("gRPC %s: %s", code, status_.Message())
		}
	} else {
		util.FailOnError(err)
	}
}

func Print(content any) {
	if !terminal.Quiet {
		Write(os.Stdout, content)
	}
}

func PrintResources(resources tkoutil.Resources) {
	if !terminal.Quiet {
		WriteResources(os.Stdout, resources)
	}
}

func Write(writer io.Writer, content any) {
	util.FailOnError(Transcriber(writer).Write(content))
}

func WriteResources(writer io.Writer, resources tkoutil.Resources) {
	content := make([]any, len(resources))
	for index, resource := range resources {
		content[index] = resource
	}
	Write(writer, content)
}

func Transcriber(writer io.Writer) *transcribe.Transcriber {
	return &transcribe.Transcriber{
		Writer:      writer,
		Format:      format,
		ForTerminal: pretty,
		Strict:      strict,
	}
}
