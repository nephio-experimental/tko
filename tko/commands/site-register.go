package commands

import (
	contextpkg "context"

	tkoutil "github.com/nephio-experimental/tko/util"
	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	siteCommand.AddCommand(siteRegisterCommand)

	siteRegisterCommand.Flags().StringToStringVarP(&siteMetadata, "metadata", "m", nil, "mergeable metadata")
	siteRegisterCommand.Flags().StringVarP(&url, "url", "u", "", "URL for mergeable package YAML manifests (can be a local directory or file)")
	siteRegisterCommand.Flags().BoolVarP(&stdin, "stdin", "i", false, "use mergeable package YAML manifests from stdin")
}

var siteRegisterCommand = &cobra.Command{
	Use:   "register [SITE ID] [[TEMPLATE ID]]",
	Short: "Register site",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), readPackageTimeout)
		util.OnExit(cancel)

		if len(args) == 2 {
			RegisterSite(context, args[0], args[1], siteMetadata, url, stdin)
		} else {
			RegisterSite(context, args[0], "", siteMetadata, url, stdin)
		}
	},
}

func RegisterSite(context contextpkg.Context, siteId string, templateId string, metadata map[string]string, url string, stdin bool) {
	var package_ tkoutil.Package
	if stdin || (url != "") {
		var err error
		package_, err = readPackage(context, url, stdin)
		util.FailOnError(err)
	}

	ok, reason, err := NewClient().RegisterSite(siteId, templateId, metadata, package_)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("registered site: %s", siteId)
	} else {
		util.Fail(reason)
	}
}
