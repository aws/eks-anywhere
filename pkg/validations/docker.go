package validations

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	recommendedTotalMemory  = 6200000000
	requiredMajorVersion    = 20
	unsupportedMinorVersion = "4.3"
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
		logger.V(3).Info("Failed to validate docker memory: error while reading memory allocated to Docker %v\n", err)
	}
	if totalMemoryAllocated < recommendedTotalMemory {
		logger.V(3).Info("Warning: recommended memory to be allocated for Docker is 6 GB, please be aware that not allocating enough memory can cause problems while cluster creation")
	}
}

func CheckDockerDesktopVersion(ctx context.Context) error {
	dockerDesktopInfoPath := "/Applications/Docker.app/Contents/Info.plist"
	if _, err := os.Stat(dockerDesktopInfoPath); err != nil {
		return fmt.Errorf("unable to find Docker Desktop info list")
	}
	cmd := exec.CommandContext(ctx, "defaults", "read", dockerDesktopInfoPath, "CFBundleShortVersionString")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}
	dockerDesktopVersion := strings.TrimSpace(string(stdout))
	dockerDesktopMinorRelease := dockerDesktopVersion[:strings.LastIndex(dockerDesktopVersion, ".")]
	if dockerDesktopMinorRelease >= unsupportedMinorVersion {
		return fmt.Errorf("EKS Anywhere does not support Docker desktop version 4.3.0 or greater on macOS")
	}

	return nil
}
