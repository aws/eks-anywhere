package cmd

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

type generatePackageOptions struct {
	directory   string
	source      string
	kubeVersion string
}

var gepo = &generatePackageOptions{}

func init() {
	generateCmd.AddCommand(generatePackageCommand)
	generatePackageCommand.Flags().StringVarP(&gepo.directory, "directory", "d", "", "Directory to save generated packages")
	generatePackageCommand.Flags().StringVar(&gepo.source, "source", "", "Location to find curated packages: (cluster, registry)")
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
		source := strings.ToLower(gepo.source)
		if err := validateSource(source); err != nil {
			return err
		}

		if err := validateKubeVersion(gepo.kubeVersion, source); err != nil {
			return err
		}

		if !validations.FileExists(gepo.directory) {
			return fmt.Errorf("directory %s does not exist", gepo.directory)
		}
		return generatePackages(cmd.Context(), gepo, args)
	}
}

func generatePackages(ctx context.Context, gepo *generatePackageOptions, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig, gepo.source, gepo.kubeVersion)
	if err != nil {
		return err
	}
	packages, err := curatedpackages.GeneratePackages(bundle, args)
	if err != nil {
		return err
	}
	err = curatedpackages.WritePackagesToFile(packages, gepo.directory)
	if err != nil {
		return err
	}
	return nil
}
