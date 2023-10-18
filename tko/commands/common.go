package commands

import (
	"io"
	"os"

	"github.com/nephio-experimental/tko/api/client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const toolName = "tko"

var log = commonlog.GetLogger(toolName)

var url string
var stdin bool
var templateIdPatterns []string
var siteIdPatterns []string
var templateMetadata map[string]string
var siteMetadata map[string]string
var parentDeploymentId string

func NewClient() *client.Client {
	client_, err := client.NewClient(grpcProtocol, grpcAddress, int(grpcPort), grpcFormat, commonlog.GetLogger("client"))
	util.FailOnError(err)
	return client_
}

func FailOnGRPCError(err error) {
	if status_, ok := status.FromError(err); ok {
		switch code := status_.Code(); code {
		case codes.OK:
			return
		case codes.Unknown:
			util.Fail(status_.Message())
		default:
			util.Failf("%s: %s", code, status_.Message())
		}
	} else {
		util.FailOnError(err)
	}
}

func Print(content any) {
	Write(os.Stdout, content)
}

func PrintResources(resources tkoutil.Resources) {
	WriteResources(os.Stdout, resources)
}

func Write(writer io.Writer, content any) {
	err := Transcriber(writer).Write(content)
	util.FailOnError(err)
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
