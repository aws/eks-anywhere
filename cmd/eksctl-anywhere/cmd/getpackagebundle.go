package cmd

import (
	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type getPackageBundleOptions struct {
	output string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	bundlesOverride string
}

var gpbo = &getPackageBundleOptions{}

func init() {
	getCmd.AddCommand(getPackageBundleCommand)

	getPackageBundleCommand.Flags().StringVarP(&gpbo.output, "output", "o", "",
		"Specifies the output format (valid option: json, yaml)")
	getPackageBundleCommand.Flags().StringVar(&gpbo.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file.")
	getPackageBundleCommand.Flags().StringVar(&gpbo.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")
}

var getPackageBundleCommand = &cobra.Command{
	Use:          "packagebundle [flags]",
	Aliases:      []string{"packagebundles"},
	Short:        "Get packagebundle(s)",
	Long:         "This command is used to display the currently supported packagebundles",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeConfig, err := kubeconfig.ResolveAndValidateFilename(gpbo.kubeConfig, "")
		if err != nil {
			return err
		}
		return getResources(cmd.Context(), "packagebundles", gpbo.output, kubeConfig, "", gpbo.bundlesOverride, args)
	},
}
