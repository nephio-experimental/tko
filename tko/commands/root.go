package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var logTo string
var verbose int
var maxWidth int
var format string
var colorize string
var strict bool
var pretty bool

var grpcIpStackString string
var grpcIpStack util.IPStack
var grpcAddress string
var grpcPort uint
var grpcFormat string

func init() {
	rootCommand.PersistentFlags().BoolVarP(&terminal.Quiet, "quiet", "q", false, "suppress output")
	rootCommand.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	rootCommand.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	rootCommand.PersistentFlags().IntVarP(&maxWidth, "width", "j", 0, "maximum output width (0 to use terminal width, -1 for no maximum)")
	rootCommand.PersistentFlags().StringVarP(&format, "format", "o", "", "output format (\"bare\", \"yaml\", \"json\", \"cjson\", \"xml\", \"cbor\", \"messagepack\", or \"go\")")
	rootCommand.PersistentFlags().StringVarP(&colorize, "colorize", "z", "true", "colorize output (boolean or \"force\")")
	rootCommand.PersistentFlags().BoolVarP(&strict, "strict", "y", false, "strict output (for \"yaml\" format only)")
	rootCommand.PersistentFlags().BoolVarP(&pretty, "pretty", "p", true, "prettify output")

	rootCommand.PersistentFlags().StringVar(&grpcIpStackString, "grpc-ip-stack", "dual", "IP stack for TKO API Server (\"dual\", \"ipv6\", or \"ipv4\")")
	rootCommand.PersistentFlags().StringVar(&grpcAddress, "grpc-address", "", "address for TKO API Server")
	rootCommand.PersistentFlags().UintVar(&grpcPort, "grpc-port", 50050, "HTTP/2 port for TKO API Server")
	rootCommand.PersistentFlags().StringVar(&grpcFormat, "grpc-format", "cbor", "preferred format for encoding resources over gRPC (\"yaml\" or \"cbor\")")
}

var rootCommand = &cobra.Command{
	Use:   toolName,
	Short: "TKO CLI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		util.InitializeColorization(colorize)
		commonlog.Initialize(verbose, logTo)

		grpcIpStack = util.IPStack(grpcIpStackString)
		util.FailOnError(grpcIpStack.Validate("grpc-ip-stack"))
	},
}

func Execute() {
	err := rootCommand.Execute()
	util.FailOnError(err)
}
