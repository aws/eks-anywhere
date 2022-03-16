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
	from        string
}

var lpo = &listPackagesOption{}

func init() {
	listCmd.AddCommand(listPackagesCommand)
	listPackagesCommand.Flags().StringVar(&lpo.from, "from", "", "Discovery Location. Options (cluster, registry)")
	listPackagesCommand.Flags().StringVar(&lpo.kubeVersion, "kubeversion", "", "Kubernetes Version of the cluster to be used. Format <major>.<minor>")
	err := listPackagesCommand.MarkFlagRequired("from")
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

		if err := packageLocationValidation(lpo.from); err != nil {
			return err
		}

		if err := kubeVersionValidation(lpo.kubeVersion, lpo.from); err != nil {
			return err
		}

		if err := listPackages(cmd.Context(), lpo.from, lpo.kubeVersion); err != nil {
			return err
		}
		return nil
	},
}

func listPackages(ctx context.Context, location string, kubeVersion string) error {
	kubeConfig := kubeconfig.FromEnvironment()
	bundle, err := curatedpackages.GetLatestBundle(ctx, kubeConfig, location, kubeVersion)
	if err != nil {
		return err
	}
	packages, err := curatedpackages.GetPackages(ctx, bundle)
	curatedpackages.DisplayPackages(packages)
	return nil
}
