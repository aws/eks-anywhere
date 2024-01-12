package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type describePackagesOption struct {
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	clusterName     string
	bundlesOverride string
}

var dpo = &describePackagesOption{}

func init() {
	describeCmd.AddCommand(describePackagesCommand)

	describePackagesCommand.Flags().StringVar(&dpo.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file to use.")
	describePackagesCommand.Flags().StringVar(&dpo.clusterName, "cluster", "",
		"Cluster to describe packages.")
	describePackagesCommand.Flags().StringVar(&dpo.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")
	if err := describePackagesCommand.MarkFlagRequired("cluster"); err != nil {
		log.Fatalf("marking cluster flag as required: %s", err)
	}
}

var describePackagesCommand = &cobra.Command{
	Use:          "package [flags]",
	Short:        "Describe curated packages in the cluster",
	Aliases:      []string{"packages"},
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
	kubeConfig, err := kubeconfig.ResolveAndValidateFilename(dpo.kubeConfig, "")
	if err != nil {
		return err
	}
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig), WithBundlesOverride(dpo.bundlesOverride))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
	)

	err = packages.DescribePackages(ctx, args, kubeConfig, dpo.clusterName)
	if err != nil {
		return err
	}

	return nil
}
