package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/constants"
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

	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl

	params := []string{"get", resourceType, "--kubeconfig", kubeConfig, "--namespace", constants.EksaPackagesName}
	params = append(params, args...)
	if output != "" {
		params = append(params, "-o", output)
	}
	stdOut, err := kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return fmt.Errorf("kubectl execution failure: \n%v", err)
	}
	if len(stdOut.Bytes()) == 0 {
		fmt.Printf("No resources found in %v namespace\n", constants.EksaPackagesName)
		return nil
	}
	fmt.Print(&stdOut)
	return nil
}
