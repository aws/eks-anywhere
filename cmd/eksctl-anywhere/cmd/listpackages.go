package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type listPackagesOption struct {
	kubeVersion string
	clusterName string
	registry    string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	bundlesOverride string
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)

	listPackagesCommand.Flags().StringVar(&lpo.kubeVersion, "kube-version", "",
		"Kubernetes version <major>.<minor> of the packages to list, for example: \"1.23\".")
	listPackagesCommand.Flags().StringVar(&lpo.registry, "registry", "",
		"Specifies an alternative registry for packages discovery.")
	listPackagesCommand.Flags().StringVar(&lpo.kubeConfig, "kubeconfig", "",
		"Path to a kubeconfig file to use when source is a cluster.")
	listPackagesCommand.Flags().StringVar(&lpo.clusterName, "cluster", "",
		"Name of cluster for package list. Required for airgapped environments.")
	listPackagesCommand.Flags().StringVar(&lpo.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")
}

var listPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Lists curated packages available to install",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := curatedpackages.ValidateKubeVersion(lpo.kubeVersion, lpo.clusterName); err != nil {
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

	depOpts := []PackageOpt{
		WithRegistryName(lpo.registry),
		WithKubeVersion(lpo.kubeVersion),
		WithMountPaths(kubeConfig),
		WithBundlesOverride(lpo.bundlesOverride),
	}
	cluster := &anywherev1.Cluster{}
	if len(lpo.clusterName) > 0 {
		k8sClient, err := kubernetes.NewRuntimeClientFromFileName(kubeConfig)
		if err != nil {
			return fmt.Errorf("unable to initalize k8s client: %v", err)
		}

		if err := k8sClient.Get(ctx, types.NamespacedName{Name: lpo.clusterName, Namespace: constants.DefaultNamespace}, cluster); err != nil {
			return fmt.Errorf("unable to get cluster %s: %v", lpo.clusterName, err)
		}
		depOpts = append(depOpts, WithCluster(cluster))
	}

	deps, err := NewDependenciesForPackages(ctx, depOpts...)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	bm := curatedpackages.CreateBundleManager(deps.Logger)

	b := curatedpackages.NewBundleReader(kubeConfig, lpo.clusterName, deps.Kubectl, bm, deps.BundleRegistry)

	bundle, err := b.GetLatestBundle(ctx, lpo.kubeVersion)
	if err != nil {
		return err
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
		curatedpackages.WithBundle(bundle),
	)
	return packages.DisplayPackages(os.Stdout)
}
