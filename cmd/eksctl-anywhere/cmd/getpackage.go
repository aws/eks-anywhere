package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
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
	Use:          "package(s) [flags]",
	Aliases:      []string{"package", "packages"},
	Short:        "Get package(s)",
	Long:         "This command is used to display the curated packages installed in the cluster",
	PreRunE:      preRunGetPackages,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getPackages(cmd.Context(), gpo.output, args)
	},
}

func preRunGetPackages(cmd *cobra.Command, args []string) error {
	if !features.IsActive(features.CuratedPackagesSupport()) {
		return fmt.Errorf("this command is currently not supported")
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err := viper.BindPFlag(flag.Name, flag); err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func getPackages(ctx context.Context, output string, args []string) error {
	kubeConfig := os.Getenv(kubeconfigEnvVariable)

	deps, err := createKubectl(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl

	params := []executables.KubectlOpt{executables.WithKubeconfig(kubeConfig), executables.WithArgs(args)}
	if output != "" {
		params = append(params, executables.WithOutput(output))
	}
	packages, err := kubectl.GetPackages(ctx, params...)
	if err != nil {
		fmt.Print(packages)
		return fmt.Errorf("error executing kubectl: %v", err)
	}
	fmt.Println(packages)
	return nil
}

func createKubectl(ctx context.Context) (*dependencies.Dependencies, error) {
	return dependencies.
		NewFactory().
		WithExecutableImage(executables.DefaultEksaImage()).
		WithExecutableBuilder().
		WithKubectl().
		Build(ctx)
}
