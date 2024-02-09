package cmd

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func commonValidation(ctx context.Context, clusterConfigFile string) (*v1alpha1.Cluster, error) {
	docker := executables.BuildDockerExecutable()
	err := validations.CheckMinimumDockerVersion(ctx, docker)
	if err != nil {
		return nil, fmt.Errorf("failed to validate docker: %v", err)
	}
	validations.CheckDockerAllocatedMemory(ctx, docker)
	clusterConfigFileExist := validations.FileExists(clusterConfigFile)
	if !clusterConfigFileExist {
		return nil, fmt.Errorf("the cluster config file %s does not exist", clusterConfigFile)
	}

	clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfigFile)
	if err != nil {
		return nil, fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}
	return clusterConfig, nil
}
