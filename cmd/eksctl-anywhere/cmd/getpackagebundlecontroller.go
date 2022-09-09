package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type getPackageBundleControllerOptions struct {
	output string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster
	kubeConfig string
}

var gpbco = &getPackageBundleControllerOptions{}

func init() {
	getCmd.AddCommand(getPackageBundleControllerCommand)

	getPackageBundleControllerCommand.Flags().StringVarP(&gpbco.output, "output",
		"o", "", "Specifies the output format (valid option: json, yaml)")
	getPackageBundleControllerCommand.Flags().StringVar(&gpbco.kubeConfig,
		"kubeconfig", "", "Path to an optional kubeconfig file.")
}

var getPackageBundleControllerCommand = &cobra.Command{
	Use:          "packagebundlecontroller(s) [flags]",
	Aliases:      []string{"packagebundlecontroller", "packagebundlcontrolleres", "pbc"},
	Short:        "Get packagebundlecontroller(s)",
	Long:         "This command is used to display the current packagebundlecontrollers",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeConfig := gpbco.kubeConfig
		if kubeConfig == "" {
			kubeConfig = kubeconfig.FromEnvironment()
		} else if !validations.FileExistsAndIsNotEmpty(kubeConfig) {
			return fmt.Errorf("kubeconfig file %q is empty or does not exist", kubeConfig)
		}
		return getResources(cmd.Context(), "packagebundlecontrollers", gpbco.output, kubeConfig, args)
	},
}
