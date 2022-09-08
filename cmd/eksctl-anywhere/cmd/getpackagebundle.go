package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type getPackageBundleOptions struct {
	output string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster
	kubeConfig string
}

var gpbo = &getPackageBundleOptions{}

func init() {
	getCmd.AddCommand(getPackageBundleCommand)

	getPackageBundleCommand.Flags().StringVarP(&gpbo.output, "output", "o", "",
		"Specifies the output format (valid option: json, yaml)")
	getPackageBundleCommand.Flags().StringVar(&gpbo.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file.")
}

var getPackageBundleCommand = &cobra.Command{
	Use:          "packagebundle(s) [flags]",
	Aliases:      []string{"packagebundle", "packagebundles"},
	Short:        "Get packagebundle(s)",
	Long:         "This command is used to display the currently supported packagebundles",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeConfig := gpbo.kubeConfig
		if kubeConfig == "" {
			kubeConfig = kubeconfig.FromEnvironment()
		} else if !validations.FileExistsAndIsNotEmpty(kubeConfig) {
			return fmt.Errorf("kubeconfig file %q is empty or does not exist", kubeConfig)
		}
		return getResources(cmd.Context(), "packagebundles", gpbo.output, kubeConfig, args)
	},
}
