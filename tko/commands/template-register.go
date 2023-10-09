package commands

import (
	contextpkg "context"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	templateCommand.AddCommand(templateRegisterCommand)

	templateRegisterCommand.Flags().StringToStringVarP(&templateMetadata, "metadata", "m", nil, "metadata")
	templateRegisterCommand.Flags().StringVarP(&url, "url", "u", "", "URL for YAML content (can be a local directory or file)")
	templateRegisterCommand.Flags().BoolVarP(&stdin, "stdin", "i", false, "use YAML content from stdin")
}

var templateRegisterCommand = &cobra.Command{
	Use:   "register [TEMPLATE ID]",
	Short: "Register template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		RegisterTemplate(contextpkg.TODO(), args[0], templateMetadata, url, stdin)
	},
}

func RegisterTemplate(context contextpkg.Context, templateId string, metadata map[string]string, url string, stdin bool) {
	resources, err := readResources(context, url, stdin)
	util.FailOnError(err)

	ok, reason, err := NewClient().RegisterTemplate(templateId, metadata, resources)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("registered template: %s", templateId)
	} else {
		util.Fail(reason)
	}
}
