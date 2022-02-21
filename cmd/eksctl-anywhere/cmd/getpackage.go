package cmd

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
)

type getPackageOptions struct {
	output string
}

var gpo = &getPackageOptions{}

func init() {
	getCmd.AddCommand(getPackageCommand)
	getPackageCommand.Flags().StringVarP(&gpo.output, "output", "o", "", "Specifies the output format (valid option: json, yaml)")
}

var getPackageCommand = &cobra.Command{
	Use:     "package(s)",
	Aliases: []string{"package", "packages"},
	Short:   "Get package(s)",
	Long:    "This command is used to display one  or more curated packages",
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
		packageInstanceName := ""
		if len(args) > 0 {
			packageInstanceName = args[0]
		}
		return getPackages(cmd.Context(), packageInstanceName, gpo.output)
	},
}

func getPackages(ctx context.Context, packageInstanceName string, output string) error {
	kubeConfig := os.Getenv(kubeconfigEnvVariable)
	executableBuilder, _, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := executableBuilder.BuildKubectlExecutable()
	packages, err := kubectl.GetPackagesFromKubectl(ctx, packageInstanceName, kubeConfig, output)
	fmt.Println(packages)
	return nil
}
