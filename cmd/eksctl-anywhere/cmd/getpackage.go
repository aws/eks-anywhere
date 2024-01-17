package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

type getPackageOptions struct {
	output string
	// kubeConfig is an optional kubeconfig file to use when querying an
	// existing cluster.
	kubeConfig      string
	clusterName     string
	bundlesOverride string
}

var gpo = &getPackageOptions{}

func init() {
	getCmd.AddCommand(getPackageCommand)

	getPackageCommand.Flags().StringVarP(&gpo.output, "output", "o", "",
		"Specifies the output format (valid option: json, yaml)")
	getPackageCommand.Flags().StringVar(&gpo.kubeConfig, "kubeconfig", "",
		"Path to an optional kubeconfig file.")
	getPackageCommand.Flags().StringVar(&gpo.clusterName, "cluster", "",
		"Cluster to get list of packages.")
	getPackageCommand.Flags().StringVar(&gpo.bundlesOverride, "bundles-override", "",
		"Override default Bundles manifest (not recommended)")
	if err := getPackageCommand.MarkFlagRequired("cluster"); err != nil {
		log.Fatalf("marking cluster flag as required: %s", err)
	}
}

var getPackageCommand = &cobra.Command{
	Use:          "package [flags]",
	Aliases:      []string{"packages"},
	Short:        "Get package(s)",
	Long:         "This command is used to display the curated packages installed in the cluster",
	PreRunE:      preRunPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		kubeConfig, err := kubeconfig.ResolveAndValidateFilename(gpo.kubeConfig, "")
		if err != nil {
			return err
		}
		return getResources(cmd.Context(), "packages", gpo.output, kubeConfig, gpo.clusterName, gpo.bundlesOverride, args)
	},
}
