package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/validations"
)

var validateClusterCmd = &cobra.Command{
	Use:   "validate-cluster <cluster-name> <kubeconfig>",
	Short: "Validate eks-a cluster command",
	Long:  "Use eks-a-tool validate eks-anywhere cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			log.Fatalf("Some args are missing. See usage for required arguments")
		}
		clusterName, err := validations.ValidateClusterNameArg(args)
		if err != nil {
			log.Fatalf("Error validating the cluster: %v", err)
		}
		kubeconfig := args[1]
		if !validations.FileExists(kubeconfig) {
			log.Fatalf("Error validating the cluster: kubeconfig file %s not found", kubeconfig)
		}
		err = validateCluster(cmd.Context(), clusterName, kubeconfig)
		if err != nil {
			log.Fatalf("Error validating the cluster: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateClusterCmd)
}

func validateCluster(ctx context.Context, clusterName string, kubeconfig string) error {
	executableBuilder, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := executableBuilder.BuildKubectlExecutable()
	err = kubectl.ValidateNodes(ctx, kubeconfig)
	if err != nil {
		return err
	}
	err = kubectl.ValidateControlPlaneNodes(ctx, clusterName, kubeconfig)
	if err != nil {
		return err
	}
	err = kubectl.ValidateWorkerNodes(ctx, clusterName, kubeconfig)
	if err != nil {
		return err
	}
	return kubectl.ValidatePods(ctx, kubeconfig)
}
