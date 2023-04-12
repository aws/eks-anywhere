package validations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	recommendedTotalMemory = 6200000000
	requiredMajorVersion   = 20
)

type DockerExecutable interface {
	Version(ctx context.Context) (int, error)
	AllocatedMemory(ctx context.Context) (uint64, error)
}

func CheckMinimumDockerVersion(ctx context.Context, dockerExecutable DockerExecutable) error {
	installedMajorVersionInt, err := dockerExecutable.Version(ctx)
	if err != nil {
		return err
	}
	if installedMajorVersionInt < requiredMajorVersion {
		return fmt.Errorf("minimum requirements for docker version have not been met. Install Docker version %d.x.x or above", requiredMajorVersion)
	}
	return nil
}

func CheckDockerAllocatedMemory(ctx context.Context, dockerExecutable DockerExecutable) {
	totalMemoryAllocated, err := dockerExecutable.AllocatedMemory(ctx)
	if err != nil {
		logger.Error(err, "Failed to validate docker memory: error while reading memory allocated to Docker")
		return
	}
	if totalMemoryAllocated < recommendedTotalMemory {
		logger.V(3).Info("Warning: recommended memory to be allocated for Docker is 6 GB, please be aware that not allocating enough memory can cause problems while cluster creation")
	}
}

func ValidateDockerExecutable(ctx context.Context, docker DockerExecutable, os string) error {
	err := CheckMinimumDockerVersion(ctx, docker)
	if err != nil {
		return fmt.Errorf("failed to validate docker: %v", err)
	}

	CheckDockerAllocatedMemory(ctx, docker)

	return nil
}
