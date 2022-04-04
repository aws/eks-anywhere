package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
	Long:  "Use eksctl anywhere get to display one or many resources",
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func preRunPackages(cmd *cobra.Command, args []string) error {
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

func getResources(ctx context.Context, resourceType string, output string, args []string) error {
	kubeConfig := kubeconfig.FromEnvironment()

	deps, err := createKubectl(ctx)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl

	params := []executables.KubectlOpt{executables.WithKubeconfig(kubeConfig), executables.WithArgs(args), executables.WithNamespace(constants.EksaPackagesName)}
	if output != "" {
		params = append(params, executables.WithOutput(output))
	}
	packages, err := kubectl.GetResources(ctx, resourceType, params...)
	if err != nil {
		fmt.Print(packages)
		return fmt.Errorf("kubectl execution failure: \n%v", err)
	}
	if packages == "" {
		fmt.Printf("No resources found in %v namespace", constants.EksaPackagesName)
		return nil
	}
	fmt.Println(packages)
	return nil
}

func createKubectl(ctx context.Context) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		UseExecutableImage(executables.DefaultEksaImage()).
		WithExecutableBuilder().
		WithKubectl().
		Build(ctx)
}
