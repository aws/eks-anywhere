package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type createPackageOptions struct {
	fileName string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	bundlesOverride string
}

var cpo = &createPackageOptions{}

func init() {
	createCmd.AddCommand(createPackagesCommand)

	createPackagesCommand.Flags().StringVarP(&cpo.fileName, "filename", "f",
		"", "Filename that contains curated packages custom resources to create")
	createPackagesCommand.Flags().StringVar(&cpo.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file to use.")
	createPackagesCommand.Flags().StringVar(&cpo.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")

	err := createPackagesCommand.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

var createPackagesCommand = &cobra.Command{
	Use:          "package [flags]",
	Short:        "Create curated packages",
	Long:         "Create Curated Packages Custom Resources to the cluster",
	Aliases:      []string{"packages"},
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := createPackages(cmd.Context()); err != nil {
			return err
		}
		return nil
	},
}

func createPackages(ctx context.Context) error {
	kubeConfig, err := kubeconfig.ResolveAndValidateFilename(cpo.kubeConfig, "")
	if err != nil {
		return err
	}
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig), WithBundlesOverride(cpo.bundlesOverride))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	packages := curatedpackages.NewPackageClient(
		deps.Kubectl,
	)

	curatedpackages.PrintLicense()
	err = packages.CreatePackages(ctx, cpo.fileName, kubeConfig)
	if err != nil {
		return err
	}

	return nil
}
