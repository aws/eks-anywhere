package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type getPackageOptions struct {
	output string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster
	kubeConfig string
}

var gpo = &getPackageOptions{}

func init() {
	getCmd.AddCommand(getPackageCommand)

	getPackageCommand.Flags().StringVarP(&gpo.output, "output", "o", "",
		"Specifies the output format (valid option: json, yaml)")
	getPackageCommand.Flags().StringVar(&gpo.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file.")
}

var getPackageCommand = &cobra.Command{
	Use:          "package(s) [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Get package(s)",
	Long:         "This command is used to display the curated packages installed in the cluster",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeConfig := gpo.kubeConfig
		if kubeConfig == "" {
			kubeConfig = kubeconfig.FromEnvironment()
		} else if !validations.FileExistsAndIsNotEmpty(kubeConfig) {
			return fmt.Errorf("kubeconfig file %q is empty or does not exist", kubeConfig)
		}
		return getResources(cmd.Context(), "packages", gpo.output, kubeConfig, args)
	},
}
