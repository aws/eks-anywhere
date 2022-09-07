package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type listPackagesOption struct {
	kubeVersion string
	source      curatedpackages.BundleSource
	registry    string
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)

	listPackagesCommand.Flags().Var(&lpo.source, "source",
		"Packages info discovery source. Options: cluster, registry.")
	listPackagesCommand.Flags().StringVar(&lpo.kubeVersion, "kube-version", "",
		"Kubernetes version of the packages to list, for example: \"1.23\".")
	listPackagesCommand.Flags().StringVar(&lpo.registry, "registry", "",
		"Specifies an alternative registry for packages discovery.")

	if err := listPackagesCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("marking source flag required: %s", err)
	}
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
	deps, err := NewDependenciesForPackages(ctx, WithRegistryName(lpo.registry), WithKubeVersion(lpo.kubeVersion), WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	bm := curatedpackages.CreateBundleManager()

	b := curatedpackages.NewBundleReader(
		kubeConfig,
		lpo.source,
		deps.Kubectl,
		bm,
		deps.BundleRegistry,
	)

	bundle, err := b.GetLatestBundle(ctx, lpo.kubeVersion)
	if err != nil {
		return err
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
		curatedpackages.WithBundle(bundle),
	)
	packages.DisplayPackages()
	return nil
}
