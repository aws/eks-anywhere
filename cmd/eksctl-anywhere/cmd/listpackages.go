package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/version"
)

type listPackagesOption struct {
	kubeVersion string
	source      curatedpackages.BundleSource
	registry    string
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)
	listPackagesCommand.Flags().Var(&lpo.source, "source", "Discovery Location. Options (cluster, registry)")
	err := listPackagesCommand.MarkFlagRequired("source")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
	listPackagesCommand.Flags().StringVar(&lpo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	listPackagesCommand.Flags().StringVar(&lpo.registry, "registry", "", "Used to specify an alternative registry for discovery")
}

var listPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Lists curated packages available to install",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := curatedpackages.ValidateKubeVersion(lpo.kubeVersion, lpo.source); err != nil {
			return err
		}

		if err := listPackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func listPackages(ctx context.Context) error {
	kubeConfig := kubeconfig.FromEnvironment()
	deps, err := curatedpackages.NewDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	bm := curatedpackages.CreateBundleManager(lpo.kubeVersion)
	username, password, err := config.ReadCredentials()
	if err != nil && gpOptions.registry != "" {
		return err
	}
	registry, err := curatedpackages.NewRegistry(deps, lpo.registry, lpo.kubeVersion, username, password)
	if err != nil {
		return err
	}

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		lpo.kubeVersion,
		lpo.source,
		deps.Kubectl,
		bm,
		version.Get(),
		registry,
	)

	bundle, err := b.GetLatestBundle(ctx)
	if err != nil {
		return err
	}
	packages := curatedpackages.NewPackageClient(
		bundle,
		deps.Kubectl,
	)
	packages.DisplayPackages()
	return nil
}
