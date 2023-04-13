package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/constants"
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
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err := viper.BindPFlag(flag.Name, flag); err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func getResources(ctx context.Context, resourceType, output, kubeConfig, clusterName, bundlesOverride string, args []string) error {
	deps, err := NewDependenciesForPackages(ctx, WithMountPaths(kubeConfig), WithBundlesOverride(bundlesOverride))
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl

	namespace := constants.EksaPackagesName
	if len(clusterName) > 0 {
		namespace = namespace + "-" + clusterName
	}
	params := []string{"get", resourceType, "--kubeconfig", kubeConfig, "--namespace", namespace}
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
