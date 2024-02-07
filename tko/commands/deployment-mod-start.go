package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/util"
)

func init() {
	deploymentModCommand.AddCommand(deploymentModStartCommand)

	deploymentModStartCommand.Flags().StringVarP(&url, "url", "u", "", "URL for YAML content output (can be a local directory or file)")
}

var deploymentModStartCommand = &cobra.Command{
	Use:   "start [DEPLOYMENT ID]",
	Short: "Start modification of a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		StartDeploymentModification(args[0], url)
	},
}

func StartDeploymentModification(deploymentId string, url string) {
	ok, reason, modificationToken, resources, err := NewClient().StartDeploymentModification(deploymentId)
	FailOnGRPCError(err)
	if ok {
		log.Noticef("started modification: %s", deploymentId)
		fmt.Println(modificationToken)
	} else {
		util.Fail(reason)
	}

	if url != "" {
		directory, err := filepath.Abs(url)
		util.FailOnError(err)
		err = os.MkdirAll(directory, 0700)
		util.FailOnError(err)
		path := filepath.Join(directory, "deployment.yaml")

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		util.FailOnError(err)
		util.OnExitError(file.Close)
		WriteResources(file, resources)
		util.FailOnError(err)
	} else {
		PrintResources(resources)
	}
}
