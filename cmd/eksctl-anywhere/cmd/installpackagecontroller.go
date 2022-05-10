package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/version"
)

type installControllerOptions struct {
	kubeVersion string
}

var ico = &installControllerOptions{}

func init() {
	installCmd.AddCommand(installPackageControllerCommand)
	installPackageControllerCommand.Flags().StringVar(&ico.kubeVersion, "kube-version", "", "Bundle version to use")
	if err := installPackageControllerCommand.MarkFlagRequired("kube-version"); err != nil {
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
	if err := curatedpackages.ValidateKubeVersion(ico.kubeVersion, curatedpackages.Registry); err != nil {
		return err
	}

	return installPackageController(cmd.Context())
}

func installPackageController(ctx context.Context) error {
	kubeConfig := kubeconfig.FromEnvironment()

	deps, err := curatedpackages.NewDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	versionBundle, err := curatedpackages.GetVersionBundle(deps.ManifestReader, version.Get().GitVersion, ico.kubeVersion)
	if err != nil {
		return err
	}
	helmChart := versionBundle.PackageController.HelmChart
	ctrlClient := curatedpackages.NewPackageControllerClient(
		deps.Helm,
		deps.Kubectl,
		kubeConfig,
		helmChart.Image(),
		helmChart.Name,
		helmChart.Tag(),
	)

	if err = ctrlClient.ValidateControllerDoesNotExist(ctx); err != nil {
		return err
	}

	curatedpackages.PrintLicense()
	err = ctrlClient.InstallController(ctx)
	if err != nil {
		return err
	}

	return nil
}
