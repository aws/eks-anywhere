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

type upgradePackageOptions struct {
	bundleVersion string
}

var upo = &upgradePackageOptions{}

func init() {
	upgradeCmd.AddCommand(upgradePackagesCommand)
	upgradePackagesCommand.Flags().StringVar(&upo.bundleVersion, "bundle-version", "", "Bundle version to use")
	err := upgradePackagesCommand.MarkFlagRequired("bundle-version")
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
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		ipo.kubeVersion,
		ipo.source,
		deps.Kubectl,
		nil,
		version.Get(),
		nil,
	)
	activeController, err := b.GetActiveController(ctx)
	if err != nil {
		return err
	}
	err = b.UpgradeBundle(ctx, activeController, upo.bundleVersion)
	if err != nil {
		return err
	}
	return nil
}
