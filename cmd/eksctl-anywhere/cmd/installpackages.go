package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/version"
)

type installPackageOptions struct {
	source      curatedpackages.BundleSource
	kubeVersion string
	packageName string
	registry    string
}

var ipo = &installPackageOptions{}

func init() {
	installCmd.AddCommand(installPackageCommand)
	installPackageCommand.Flags().Var(&ipo.source, "source", "Location to find curated packages: (cluster, registry)")
	if err := installPackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	installPackageCommand.Flags().StringVar(&ipo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	installPackageCommand.Flags().StringVarP(&ipo.packageName, "packagename", "p", "", "Custom name of the curated package to install")
	if err := installPackageCommand.MarkFlagRequired("packagename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var installPackageCommand = &cobra.Command{
	Use:          "packages [flags]",
	Aliases:      []string{"package"},
	Short:        "Install package(s)",
	Long:         "This command is used to Install a curated package. Use list to discover curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runInstallPackages,
}

func runInstallPackages(cmd *cobra.Command, args []string) error {
	if err := curatedpackages.ValidateKubeVersion(ipo.kubeVersion, ipo.source); err != nil {
		return err
	}

	return installPackages(cmd.Context(), args)
}

func installPackages(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := newDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	bm := curatedpackages.CreateBundleManager(ipo.kubeVersion)
	username, password, err := config.ReadCredentials()
	if err != nil && gpOptions.registry != "" {
		return err
	}
	registry, err := curatedpackages.NewRegistry(deps, ipo.registry, ipo.kubeVersion, username, password)
	if err != nil {
		return err
	}

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		ipo.kubeVersion,
		ipo.source,
		deps.Kubectl,
		bm,
		version.Get(),
		registry,
	)

	bundle, err := b.GetLatestBundle(ctx)
	if err != nil {
		return err
	}

	packages := curatedpackages.NewPackageClient(
		bundle,
		deps.Kubectl,
	)

	p, err := packages.GetPackageFromBundle(args[0])
	if err != nil {
		return err
	}
	err = packages.InstallPackage(ctx, p, ipo.packageName, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}
