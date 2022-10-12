package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/setupuser"
)

type vSphereSetupUserOptions struct {
	fileName string
	force    bool
	password string
}

var setupUserOptions = &vSphereSetupUserOptions{}

var setupUserCmd = &cobra.Command{
	Use:          "user -f <config-file> [flags]",
	Short:        "Setup vSphere user",
	Long:         "Use eksctl anywhere vsphere setup user to configure EKS Anywhere vSphere user",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: false,
	RunE:         setupUserOptions.setupUser,
}

func init() {
	vsphereSetupCmd.AddCommand(setupUserCmd)

	setupUserCmd.Flags().StringVarP(&setupUserOptions.fileName, "filename", "f", "", "Filename containing vsphere setup configuration")
	setupUserCmd.Flags().StringVarP(&setupUserOptions.password, "password", "p", "", "Password for creating new user")
	setupUserCmd.Flags().BoolVarP(&setupUserOptions.force, "force", "", false, "Force flag. When set, setup user will proceed even if group and role objects already exists. default: false")

	if err := setupUserCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("error marking flag as required: %v", err)
	}
}

func (setupUserOptions *vSphereSetupUserOptions) setupUser(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	if setupUserOptions.force && setupUserOptions.password != "" {
		return fmt.Errorf("--password and --force are mutually exclusive. --force may only be run on an existing user")
	}

	cfg, err := setupuser.GenerateConfig(ctx, setupUserOptions.fileName)
	if err != nil {
		return err
	}

	err = setupuser.SetupGOVCEnv(ctx, cfg)
	if err != nil {
		return err
	}
	deps, err := dependencies.NewFactory().WithGovc().Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	// if force flag not used, we should create a user
	if !setupUserOptions.force {
		err = deps.Govc.CreateUser(ctx, cfg.Spec.Username, setupUserOptions.password)
	}
	if err != nil {
		return err
	}

	if !setupUserOptions.force {
		err = setupuser.ValidateVSphereObjects(ctx, cfg, deps.Govc)
		if err != nil {
			return err
		}
	}

	err = setupuser.Run(ctx, cfg, deps.Govc)
	if err != nil {
		return err
	}

	return nil
}
