package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/commonlog"
	"github.com/tliron/commonlog/journal"
	"github.com/tliron/commonlog/simple"
	"github.com/tliron/commonlog/sink"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var (
	logTo    string
	verbose  int
	colorize string
	journald bool
)

func init() {
	rootCommand.PersistentFlags().BoolVarP(&terminal.Quiet, "quiet", "q", false, "suppress output")
	rootCommand.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	rootCommand.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	rootCommand.PersistentFlags().StringVarP(&colorize, "colorize", "z", "true", "colorize output (boolean or \"force\")")
	rootCommand.PersistentFlags().BoolVarP(&journald, "journald", "j", false, "use journald logging")
}

var rootCommand = &cobra.Command{
	Use:   toolName,
	Short: "TKO Data",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		util.InitializeColorization(colorize)
		if journald {
			backend := journal.NewBackend()
			backend.VarsInMessage = true
			commonlog.SetBackend(backend)
		} else {
			commonlog.SetBackend(simple.NewBackend())
		}
		commonlog.Initialize(verbose, logTo)
		sink.CaptureKlogOutput("klog")
	},
}

func Execute() {
	err := rootCommand.Execute()
	util.FailOnError(err)
}
