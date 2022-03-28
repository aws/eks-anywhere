package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/dependencies"
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
	deps, err := newDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	bm := createBundleManager(lpo.kubeVersion)
	registry, err := newRegistry(ctx, deps, lpo)
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
		bundle.Spec.Packages,
	)
	packages.DisplayPackages()
	return nil
}

func createBundleManager(kubeVersion string) bundle.Manager {
	versionSplit := strings.Split(kubeVersion, ".")
	if len(versionSplit) != 2 {
		return nil
	}
	major, minor := versionSplit[0], versionSplit[1]
	log := logr.Discard()
	discovery := curatedpackages.NewDiscovery(major, minor)
	puller := artifacts.NewRegistryPuller()
	return bundle.NewBundleManager(log, discovery, puller)
}

func newRegistry(ctx context.Context, deps *dependencies.Dependencies, lpo *listPackagesOption) (curatedpackages.BundleRegistry, error) {
	if lpo.registry != "" {
		registryUsername := os.Getenv("REGISTRY_USERNAME")
		registryPassword := os.Getenv("REGISTRY_PASSWORD")
		if registryUsername == "" || registryPassword == "" {
			return nil, fmt.Errorf("username or password not set. Provide REGISTRY_USERNAME and REGISTRY_PASSWORD when using custom registry")
		}
		registry := curatedpackages.NewCustomRegistry(
			deps.Helm,
			lpo.registry,
			registryUsername,
			registryPassword,
		)
		err := registry.Login(ctx)
		if err != nil {
			return nil, err
		}
		return registry, nil
	}
	defaultRegistry := curatedpackages.NewDefaultRegistry(
		deps.ManifestReader,
		lpo.kubeVersion,
		version.Get(),
	)
	return defaultRegistry, nil
}
