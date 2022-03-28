package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"log"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type listPackagesOption struct {
	kubeVersion string
	source      curatedpackages.BundleSource
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)
	listPackagesCommand.Flags().Var(&lpo.source, "source", "Discovery Location. Options (cluster, registry)")
	listPackagesCommand.Flags().StringVar(&lpo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	err := listPackagesCommand.MarkFlagRequired("source")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var listPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Lists curated packages available to install",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := validateKubeVersion(lpo.kubeVersion, lpo.source); err != nil {
			return err
		}

		if err := listPackages(cmd.Context(), lpo.source, lpo.kubeVersion); err != nil {
			return err
		}
		return nil
	},
}

func listPackages(ctx context.Context, source curatedpackages.BundleSource, kubeVersion string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig, source, kubeVersion)
	if err != nil {
		return err
	}
	packages := bundle.Spec.Packages
	curatedpackages.DisplayPackages(packages)
	return nil
}
