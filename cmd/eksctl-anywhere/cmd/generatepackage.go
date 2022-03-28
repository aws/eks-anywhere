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

var gepo = &generatePackageOptions{}

func init() {
	generateCmd.AddCommand(generatePackageCommand)
	generatePackageCommand.Flags().StringVarP(&gepo.directory, "directory", "d", "", "Directory to save generated packages")
	generatePackageCommand.Flags().Var(&gepo.source, "source", "Location to find curated packages: (cluster, registry)")
	if err := generatePackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	generatePackageCommand.Flags().StringVar(&gepo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	if err := generatePackageCommand.MarkFlagRequired("directory"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	generatePackageCommand.Flags().StringVar(&gepo.registry, "registry", "", "Used to specify an alternative registry for package generation")
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
		if err := curatedpackages.ValidateKubeVersion(gepo.kubeVersion, gepo.source); err != nil {
			return err
		}
		if !validations.FileExists(gepo.directory) {
			return fmt.Errorf("directory %s does not exist", gepo.directory)
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
	bm := curatedpackages.CreateBundleManager(gepo.kubeVersion)
	registry, err := curatedpackages.NewRegistry(ctx, deps, gepo.registry, gepo.kubeVersion)
	if err != nil {
		return err
	}

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		gepo.kubeVersion,
		gepo.source,
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
	if err = packageClient.WritePackagesToFile(packages, gepo.directory); err != nil {
		return err
	}
	return nil
}
