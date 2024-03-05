package commands

import (
	"io"
	"os"
	"strings"
	"time"

	clientpkg "github.com/nephio-experimental/tko/api/grpc-client"
	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-transcribe"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

const toolName = "tko"

var (
	log                = commonlog.GetLogger(toolName)
	clientLog          = commonlog.NewScopeLogger(log, "client")
	readPackageTimeout = 10 * time.Second

	url                string
	stdin              bool
	templateIdPatterns []string
	siteIdPatterns     []string
	templateMetadata   map[string]string
	siteMetadata       map[string]string
	deploymentMetadata map[string]string
	parentDeploymentId string
	executor           string
	offset             uint
	maxCount           uint
)

func NewClient() *clientpkg.Client {
	client := clientpkg.NewClient(grpcIpStack, grpcAddress, int(grpcPort), grpcFormat, tkoutil.SecondsToDuration(grpcTimeout), clientLog)
	return client
}

func FailOnGRPCError(err error) {
	if status, ok := statuspkg.FromError(err); ok {
		switch code := status.Code(); code {
		case codes.OK:
			return
		case codes.Unknown:
			util.Fail(status.Message())
		default:
			util.Failf("gRPC %s: %s", code, status.Message())
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

func PrintPackage(package_ tkoutil.Package) {
	if !terminal.Quiet {
		WritePackage(os.Stdout, package_)
	}
}

func Write(writer io.Writer, content any) {
	util.FailOnError(Transcriber(writer).Write(content))
}

func WritePackage(writer io.Writer, package_ tkoutil.Package) {
	// This will transcribe as multiple documents in YAML
	content := make([]any, len(package_))
	for index, resource := range package_ {
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

func ParseTrigger(trigger string) *tkoutil.GVK {
	if trigger != "" {
		if s := strings.Split(trigger, ","); len(s) == 3 {
			gvk := tkoutil.NewGVK(s[0], s[1], s[2])
			return &gvk
		} else {
			util.Failf("invalid \"--trigger\": %s", trigger)
		}
	}

	return nil
}
