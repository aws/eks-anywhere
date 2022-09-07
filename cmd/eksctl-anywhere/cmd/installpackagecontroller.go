package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
)

type installControllerOptions struct {
	fileName string
}

var ico = &installControllerOptions{}

func init() {
	installCmd.AddCommand(installPackageControllerCommand)
	installPackageControllerCommand.Flags().StringVarP(&ico.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	if err := installPackageControllerCommand.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var installPackageControllerCommand = &cobra.Command{
	Use:          "packagecontroller",
	Aliases:      []string{"pc"},
	Short:        "Install packagecontroller on the cluster",
	Long:         "This command is used to Install the packagecontroller on to an existing cluster",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runInstallPackageController,
}

func runInstallPackageController(cmd *cobra.Command, args []string) error {
	clusterConfigFileExist := validations.FileExists(ico.fileName)
	if !clusterConfigFileExist {
		return fmt.Errorf("the cluster config file %s does not exist", ico.fileName)
	}
	return installPackageController(cmd.Context())
}

func installPackageController(ctx context.Context) error {
	kubeConfig := kubeconfig.FromEnvironment()

	clusterSpec, err := readAndValidateClusterSpec(ico.fileName, version.Get())
	if err != nil {
		return fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}

	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig), WithClusterSpec(clusterSpec))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	ctrlClient := deps.PackageControllerClient

	if err = curatedpackages.VerifyCertManagerExists(ctx, deps.Kubectl, kubeConfig); err != nil {
		return err
	}

	if ctrlClient.IsInstalled(ctx) {
		return errors.New("curated Packages controller exists in the current cluster")
	}

	curatedpackages.PrintLicense()
	err = ctrlClient.InstallController(ctx)
	if err != nil {
		return err
	}

	return nil
}
