package commands

import (
	contextpkg "context"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateRegisterCommand)

	templateRegisterCommand.Flags().StringToStringVarP(&templateMetadata, "metadata", "m", nil, "metadata")
	templateRegisterCommand.Flags().StringVarP(&url, "url", "u", "", "URL for package YAML manifests (can be a local directory or file)")
	templateRegisterCommand.Flags().BoolVarP(&stdin, "stdin", "i", false, "read package YAML manifests from stdin")
}

var templateRegisterCommand = &cobra.Command{
	Use:   "register [TEMPLATE ID]",
	Short: "Register template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		context, cancel := contextpkg.WithTimeout(contextpkg.Background(), readPackageTimeout)
		util.OnExit(cancel)

		RegisterTemplate(context, args[0], templateMetadata, url, stdin)
	},
}

func RegisterTemplate(context contextpkg.Context, templateId string, metadata map[string]string, url string, stdin bool) {
	package_, err := readPackage(context, url, stdin)
	util.FailOnError(err)

	ok, reason, err := NewClient().RegisterTemplate(templateId, metadata, package_)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("registered template: %s", templateId)
	} else {
		util.Fail(reason)
	}
}
