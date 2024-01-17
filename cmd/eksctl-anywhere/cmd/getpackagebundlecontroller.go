package cmd

import (
	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type getPackageBundleControllerOptions struct {
	output string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	bundlesOverride string
}

var gpbco = &getPackageBundleControllerOptions{}

func init() {
	getCmd.AddCommand(getPackageBundleControllerCommand)

	getPackageBundleControllerCommand.Flags().StringVarP(&gpbco.output, "output",
		"o", "", "Specifies the output format (valid option: json, yaml)")
	getPackageBundleControllerCommand.Flags().StringVar(&gpbco.kubeConfig,
		"kubeconfig", "", "Path to an optional kubeconfig file.")
	getPackageBundleControllerCommand.Flags().StringVar(&gpbco.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")
}

var getPackageBundleControllerCommand = &cobra.Command{
	Use:          "packagebundlecontroller [flags]",
	Aliases:      []string{"packagebundlecontrollers", "pbc"},
	Short:        "Get packagebundlecontroller(s)",
	Long:         "This command is used to display the current packagebundlecontrollers",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeConfig, err := kubeconfig.ResolveAndValidateFilename(gpbco.kubeConfig, "")
		if err != nil {
			return err
		}
		return getResources(cmd.Context(), "packagebundlecontrollers", gpbco.output, kubeConfig, "", gpbco.bundlesOverride, args)
	},
}
