package cmd

import (
	"context"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/spf13/cobra"
	"log"
)

type listPackagesOption struct {
	kubeVersion string
	source      string
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)
	listPackagesCommand.Flags().StringVar(&lpo.source, "source", "", "Discovery Location. Options (cluster, registry)")
	listPackagesCommand.Flags().StringVar(&lpo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	err := listPackagesCommand.MarkFlagRequired("source")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var listPackagesCommand = &cobra.Command{
	Use:          "packages",
	Short:        "Generate a list of curated packages available to install",
	Long:         "This command is used to generate a list of curated packages available to install",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := packageLocationValidation(lpo.source); err != nil {
			return err
		}

		if err := kubeVersionValidation(lpo.kubeVersion, lpo.source); err != nil {
			return err
		}

		if err := listPackages(cmd.Context(), lpo.source, lpo.kubeVersion); err != nil {
			return err
		}
		return nil
	},
}

func listPackages(ctx context.Context, source string, kubeVersion string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig, source, kubeVersion)
	if err != nil {
		return err
	}
	packages, err := curatedpackages.GetPackages(ctx, bundle)
	curatedpackages.DisplayPackages(packages)
	return nil
}
