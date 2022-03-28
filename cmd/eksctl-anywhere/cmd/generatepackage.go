package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type generatePackageOptions struct {
	directory   string
	source      curatedpackages.BundleSource
	kubeVersion string
}

var gepo = &generatePackageOptions{}

func init() {
	generateCmd.AddCommand(generatePackageCommand)
	generatePackageCommand.Flags().StringVarP(&gepo.directory, "directory", "d", "", "Directory to save generated packages")
	generatePackageCommand.Flags().Var(&gepo.source, "source", "Location to find curated packages: (cluster, registry)")
	generatePackageCommand.Flags().StringVar(&gepo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")

	if err := generatePackageCommand.MarkFlagRequired("directory"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	if err := generatePackageCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
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
		if err := validateKubeVersion(gepo.kubeVersion, gepo.source); err != nil {
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
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig, gepo.source, gepo.kubeVersion)
	if err != nil {
		return err
	}
	packages, err := curatedpackages.GeneratePackages(bundle, args)
	if err != nil {
		return err
	}
	if err = curatedpackages.WritePackagesToFile(packages, gepo.directory); err != nil {
		return err
	}
	return nil
}
