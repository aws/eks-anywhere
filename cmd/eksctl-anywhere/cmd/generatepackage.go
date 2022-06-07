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

type generatePackageOptions struct {
	source      curatedpackages.BundleSource
	kubeVersion string
	registry    string
}

var gpOptions = &generatePackageOptions{}

func init() {
	generateCmd.AddCommand(generatePackageCommand)
	generatePackageCommand.Flags().Var(&gpOptions.source, "source", "Location to find curated packages: (cluster, registry)")
	if err := generatePackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	generatePackageCommand.Flags().StringVar(&gpOptions.kubeVersion, "kube-version", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	generatePackageCommand.Flags().StringVar(&gpOptions.registry, "registry", "", "Used to specify an alternative registry for package generation")
}

var generatePackageCommand = &cobra.Command{
	Use:          "packages [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Generate package(s) configuration",
	Long:         "Generates Kubernetes configuration files for curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runGeneratePackages,
	Args:         cobra.MinimumNArgs(1),
}

func runGeneratePackages(cmd *cobra.Command, args []string) error {
	if err := curatedpackages.ValidateKubeVersion(gpOptions.kubeVersion, gpOptions.source); err != nil {
		return err
	}
	return generatePackages(cmd.Context(), args)
}

func generatePackages(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := NewDependenciesForPackages(ctx, WithRegistryName(gpOptions.registry), WithKubeVersion(gpOptions.kubeVersion), WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	bm := curatedpackages.CreateBundleManager(gpOptions.kubeVersion)

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		gpOptions.kubeVersion,
		gpOptions.source,
		deps.Kubectl,
		bm,
		version.Get(),
		deps.BundleRegistry,
	)

	bundle, err := b.GetLatestBundle(ctx)
	if err != nil {
		return err
	}

	packageClient := curatedpackages.NewPackageClient(
		deps.Kubectl,
		curatedpackages.WithBundle(bundle),
		curatedpackages.WithCustomPackages(args),
	)
	packages, err := packageClient.GeneratePackages()
	if err != nil {
		return err
	}
	if err = packageClient.WritePackagesToStdOut(packages); err != nil {
		return err
	}
	return nil
}
