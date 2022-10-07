package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type listPackagesOption struct {
	kubeVersion string
	clusterName string
	source      curatedpackages.BundleSource
	registry    string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig string
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)

	listPackagesCommand.Flags().Var(&lpo.source, "source",
		"Packages discovery source. Options: cluster, registry.")
	listPackagesCommand.Flags().StringVar(&lpo.kubeVersion, "kube-version", "",
		"Kubernetes version <major>.<minor> of the packages to list, for example: \"1.23\".")
	listPackagesCommand.Flags().StringVar(&lpo.registry, "registry", "",
		"Specifies an alternative registry for packages discovery.")
	listPackagesCommand.Flags().StringVar(&lpo.kubeConfig, "kubeconfig", "",
		"Path to a kubeconfig file to use when source is a cluster.")
	listPackagesCommand.Flags().StringVar(&lpo.clusterName, "cluster", "",
		"Name of cluster for package list.")

	if err := listPackagesCommand.MarkFlagRequired("source"); err != nil {
		log.Fatalf("marking source flag required: %s", err)
	}
	if err := listPackagesCommand.MarkFlagRequired("cluster"); err != nil {
		log.Fatalf("cluster flag required: %s", err)
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
	kubeConfig, err := kubeconfig.ResolveAndValidateFilename(lpo.kubeConfig, "")
	if err != nil {
		return err
	}

	opts := curatedpackages.BundleClientOptions{
		ClusterName: lpo.clusterName,
		KubeConfig:  kubeConfig,
		KubeVersion: lpo.kubeVersion,
		Registry:    lpo.registry,
	}
	bc, err := curatedpackages.NewBundleClient(lpo.source, opts)
	if err != nil {
		return err
	}
	bundle, err := bc.ActiveOrLatest(ctx)
	if err != nil {
		return err
	}

	deps, err := NewDependenciesForPackages(ctx, WithRegistryName(lpo.registry), WithKubeVersion(lpo.kubeVersion), WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
		curatedpackages.WithBundle(bundle),
	)
	return packages.DisplayPackages(os.Stdout)
}
