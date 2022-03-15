package cmd

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	FromCluster  = "cluster"
	FromRegistry = "registry"
)

func commonValidation(ctx context.Context, clusterConfigFile string) (*v1alpha1.Cluster, error) {
	docker := executables.BuildDockerExecutable()
	err := validations.CheckMinimumDockerVersion(ctx, docker)
	if err != nil {
		return nil, fmt.Errorf("failed to validate docker: %v", err)
	}
	if runtime.GOOS == "darwin" {
		err = validations.CheckDockerDesktopVersion(ctx, docker)
		if err != nil {
			return nil, fmt.Errorf("failed to validate docker desktop: %v", err)
		}
	}
	validations.CheckDockerAllocatedMemory(ctx, docker)
	clusterConfigFileExist := validations.FileExists(clusterConfigFile)
	if !clusterConfigFileExist {
		return nil, fmt.Errorf("the cluster config file %s does not exist", clusterConfigFile)
	}
	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(clusterConfigFile)
	if err != nil {
		return nil, fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}
	return clusterConfig, nil
}

func packageLocationValidation(location string) error {
	switch strings.ToLower(location) {
	case FromCluster:
	case FromRegistry:
		return nil
	}
	return fmt.Errorf("invalid from flag specified. Please use either %v, or %v", FromCluster, FromRegistry)
}

func kubeVersionValidation(kubeVersion string) error {
	versionSplit := strings.Split(kubeVersion, ".")
	if len(versionSplit) < 2 {
		return fmt.Errorf("please specify kubeVersion as <major>.<minor>")
	}
	return nil
}
