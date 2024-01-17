package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type applyPackageOptions struct {
	fileName string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	bundlesOverride string
}

var apo = &applyPackageOptions{}

func init() {
	applyCmd.AddCommand(applyPackagesCommand)

	applyPackagesCommand.Flags().StringVarP(&apo.fileName, "filename", "f",
		"", "Filename that contains curated packages custom resources to apply")
	applyPackagesCommand.Flags().StringVar(&apo.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file to use.")
	applyPackagesCommand.Flags().StringVar(&apo.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")

	err := applyPackagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var applyPackagesCommand = &cobra.Command{
	Use:          "package [flags]",
	Short:        "Apply curated packages",
	Long:         "Apply Curated Packages Custom Resources to the cluster",
	Aliases:      []string{"packages"},
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := applyPackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func applyPackages(ctx context.Context) error {
	kubeConfig, err := kubeconfig.ResolveAndValidateFilename(apo.kubeConfig, "")
	if err != nil {
		return err
	}

	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig), WithBundlesOverride(apo.bundlesOverride))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
	)

	curatedpackages.PrintLicense()
	err = packages.ApplyPackages(ctx, apo.fileName, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}
