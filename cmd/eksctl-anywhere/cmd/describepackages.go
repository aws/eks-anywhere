package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

func init() {
	describeCmd.AddCommand(describePackagesCommand)
}

var describePackagesCommand = &cobra.Command{
	Use:          "package(s) [flags]",
	Short:        "Describe curated packages in the cluster",
	Aliases:      []string{"package", "packages"},
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := describeResources(cmd.Context(), args); err != nil {
			return err
		}
		return nil
	},
}

func describeResources(ctx context.Context, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
	)

	err = packages.DescribePackages(ctx, args, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}
