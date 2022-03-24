package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type applyPackageOptions struct {
	fileName string
}

var apo = &applyPackageOptions{}

func init() {
	applyCmd.AddCommand(applyPackagesCommand)
	applyPackagesCommand.Flags().StringVarP(&apo.fileName, "filename", "f", "", "File with curated packages to apply")
	err := applyPackagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var applyPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Apply curated packages",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := applyPackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func applyPackages(ctx context.Context) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := newDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		nil,
		deps.Kubectl,
	)

	err = packages.ApplyPackages(ctx, apo.fileName, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}
