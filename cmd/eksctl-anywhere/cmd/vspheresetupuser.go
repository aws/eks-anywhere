package cmd

import (
	"fmt"
	"log"
	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser"
)

type vSphereSetupUserOptions struct {
	clusterOptions
	timeoutOptions
	fileName            string
	force           bool
	password       string
	tinkerbellBootstrapIP string
	installPackages       string
}

var vsuo = &vSphereSetupUserOptions{}


var setupUserCmd = &cobra.Command{
	Use:   "user -f <config-file> [flags]",
	Short: "Setup vSphere user",
	Long:  "Use eksctl anywhere vsphere setup user to configure EKS Anywhere vSphere user",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: false,
	RunE:         vsuo.setupUser,
}

func init() {
	vsphereSetupCmd.AddCommand(setupUserCmd)

	setupUserCmd.Flags().StringVarP(&vsuo.fileName, "filename", "f", "", "Filename containing vsphere setup configuration")
	setupUserCmd.Flags().StringVarP(&vsuo.password, "password", "p", "", "Password for creating new user")
	setupUserCmd.Flags().BoolVarP(&vsuo.force, "force", "", false, "default: false")

	if err := setupUserCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}


func (vsuo *vSphereSetupUserOptions) setupUser(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	if vsuo.force && vsuo.password != "" {
		return fmt.Errorf("--password and --force are mutually exclusive. --force may only be run on an existing user.")
	}

	deps, err := dependencies.NewFactory().WithGovc().Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	c, err := setupuser.GenerateConfig(ctx, vsuo.fileName)
	err = setupuser.SetupUser(ctx, c, deps)

	return nil
}