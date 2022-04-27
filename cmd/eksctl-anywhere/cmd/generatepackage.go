package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/version"
)

type generatePackageOptions struct {
	directory   string
	source      curatedpackages.BundleSource
	kubeVersion string
	registry    string
}

var gpOptions = &generatePackageOptions{}

func init() {
	generateCmd.AddCommand(generatePackageCommand)
	generatePackageCommand.Flags().StringVarP(&gpOptions.directory, "directory", "d", "", "Directory to save generated packages")
	generatePackageCommand.Flags().Var(&gpOptions.source, "source", "Location to find curated packages: (cluster, registry)")
	if err := generatePackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	generatePackageCommand.Flags().StringVar(&gpOptions.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	if err := generatePackageCommand.MarkFlagRequired("directory"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	generatePackageCommand.Flags().StringVar(&gpOptions.registry, "registry", "", "Used to specify an alternative registry for package generation")
}

var generatePackageCommand = &cobra.Command{
	Use:          "packages [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Generate package(s)",
	Long:         "This command is used to generate curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE:         runGeneratePackages(),
}

func runGeneratePackages() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := curatedpackages.ValidateKubeVersion(gpOptions.kubeVersion, gpOptions.source); err != nil {
			return err
		}
		if !validations.FileExists(gpOptions.directory) {
			return fmt.Errorf("directory %s does not exist", gpOptions.directory)
		}
		return generatePackages(cmd.Context(), args)
	}
}

func generatePackages(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := newDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	bm := curatedpackages.CreateBundleManager(gpOptions.kubeVersion)
	registry, err := curatedpackages.NewRegistry(ctx, deps, gpOptions.registry, gpOptions.kubeVersion)
	if err != nil {
		return err
	}

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		gpOptions.kubeVersion,
		gpOptions.source,
		deps.Kubectl,
		bm,
		version.Get(),
		registry,
	)

	bundle, err := b.GetLatestBundle(ctx)
	if err != nil {
		return err
	}

	packageClient := curatedpackages.NewPackageClient(
		bundle,
		args...,
	)
	packages, err := packageClient.GeneratePackages()
	if err != nil {
		return err
	}
	if err = packageClient.WritePackagesToFile(packages, gpOptions.directory); err != nil {
		return err
	}
	return nil
}
