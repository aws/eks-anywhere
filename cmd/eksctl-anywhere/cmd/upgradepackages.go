package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type upgradePackageOptions struct {
	bundleVersion string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	clusterName     string
	bundlesOverride string
}

var upo = &upgradePackageOptions{}

func init() {
	upgradeCmd.AddCommand(upgradePackagesCommand)

	upgradePackagesCommand.Flags().StringVar(&upo.bundleVersion, "bundle-version",
		"", "Bundle version to use")
	upgradePackagesCommand.Flags().StringVar(&upo.kubeConfig, "kubeconfig",
		"", "Path to an optional kubeconfig file to use.")
	upgradePackagesCommand.Flags().StringVar(&upo.clusterName, "cluster",
		"", "Cluster to upgrade.")
	upgradePackagesCommand.Flags().StringVar(&upo.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")

	err := upgradePackagesCommand.MarkFlagRequired("bundle-version")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	err = upgradePackagesCommand.MarkFlagRequired("cluster")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var upgradePackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Upgrade all curated packages to the latest version",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := upgradePackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func upgradePackages(ctx context.Context) error {
	kubeConfig, err := kubeconfig.ResolveAndValidateFilename(upo.kubeConfig, "")
	if err != nil {
		return err
	}

	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig), WithBundlesOverride(upo.bundlesOverride))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	b := curatedpackages.NewBundleReader(kubeConfig, upo.clusterName, deps.Kubectl, nil, nil)
	activeController, err := b.GetActiveController(ctx)
	if err != nil {
		return err
	}
	return b.UpgradeBundle(ctx, activeController, upo.bundleVersion)
}
