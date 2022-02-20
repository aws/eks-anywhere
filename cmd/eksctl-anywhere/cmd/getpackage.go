package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
)

type getPackageOptions struct {
	name   string
	output string
}

var gpo = &getPackageOptions{}

func init() {
	getCmd.AddCommand(getPackageCommand)
	getPackageCommand.Flags().StringVar(&gpo.name, "name", "", "Package Name")
	getPackageCommand.Flags().StringVarP(&gpo.name, "output", "o", "", "Specifies the output format (valid option: table, json, yaml) (default table)")
}

var getPackageCommand = &cobra.Command{
	Use:     "package(s)",
	Aliases: []string{"package", "packages"},
	Short:   "Get package(s)",
	Long:    "This command is used to display one  or many curated packages",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if err := viper.BindPFlag(flag.Name, flag); err != nil {
				log.Fatalf("Error initializing flags: %v", err)
			}
		})
		return nil
	},
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getPackages(cmd.Context(), gpo.name, gpo.output)
	},
}

func getPackages(context context.Context, name string, output string) error {
	//var kube_config = "/Users/acool/Desktop/Amazon/eks-anywhere-cluster/local-v8-cluster/local-v8-cluster-eks-a-cluster.kubeconfig"
	//getPackagesFromKubectl(context, "test", "test_ns", kube_config, "yaml")
	return nil
}
