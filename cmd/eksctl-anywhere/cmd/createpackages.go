package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type createPackageOptions struct {
	fileName string
}

var cpo = &createPackageOptions{}

func init() {
	createCmd.AddCommand(createPackagesCommand)
	createPackagesCommand.Flags().StringVarP(&cpo.fileName, "filename", "f", "", "Filename that contains curated packages custom resources to create")
	err := createPackagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var createPackagesCommand = &cobra.Command{
	Use:          "package(s) [flags]",
	Short:        "Create curated packages",
	Long:         "Create Curated Packages Custom Resources to the cluster",
	Aliases:      []string{"package", "packages"},
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := createPackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func createPackages(ctx context.Context) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
	)

	curatedpackages.PrintLicense()
	err = packages.CreatePackages(ctx, cpo.fileName, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}
