package validations

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/semver"
)

const (
	recommendedTotalMemory         = 6200000000
	requiredMajorVersion           = 20
	minUnsupportedVersion          = "4.3.0"
	minSupportedWithSettingVersion = "4.4.2"
)

type DockerExecutable interface {
	Version(ctx context.Context) (int, error)
	AllocatedMemory(ctx context.Context) (uint64, error)
	CgroupVersion(ctx context.Context) (int, error)
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

func CheckDockerDesktopVersion(ctx context.Context, dockerExecutable DockerExecutable) error {
	dockerDesktopInfoPath := "/Applications/Docker.app/Contents/Info.plist"
	if _, err := os.Stat(dockerDesktopInfoPath); err != nil {
		return fmt.Errorf("unable to find Docker Desktop info list")
	}
	cmd := exec.CommandContext(ctx, "defaults", "read", dockerDesktopInfoPath, "CFBundleShortVersionString")
	stdout, err := cmd.Output()
	if err != nil {
		return err
	}

	return ValidateDockerDesktopVersion(ctx, dockerExecutable, string(stdout))
}

func ValidateDockerDesktopVersion(ctx context.Context, dockerExecutable DockerExecutable, dockerDesktopVersion string) error {
	minUnsupportedSemVer, err := semver.New(minUnsupportedVersion)
	if err != nil {
		return err
	}

	minSupportWithSettingSemVer, err := semver.New(minSupportedWithSettingVersion)
	if err != nil {
		return err
	}

	dockerDesktopSemVer, err := semver.New(strings.TrimSpace(dockerDesktopVersion))
	if err != nil {
		return err
	}

	if dockerDesktopSemVer.LessThan(minUnsupportedSemVer) {
		// Older versions of docker desktop are supported as is
		return nil
	}

	if dockerDesktopSemVer.LessThan(minSupportWithSettingSemVer) {
		return fmt.Errorf("EKS Anywhere does not support Docker desktop versions between 4.3.0 and 4.4.1 on macOS, please refer to https://github.com/aws/eks-anywhere/issues/789 for more information")
	}

	cgroupVersion, err := dockerExecutable.CgroupVersion(ctx)
	if err != nil {
		return err
	}

	if cgroupVersion != 1 {
		return fmt.Errorf("EKS Anywhere requires Docker desktop to be configured to use CGroups v1. " +
			"Please  set `deprecatedCgroupv1:true` in your `~/Library/Group\\ Containers/group.com.docker/settings.json` file")
	}

	return nil
}

func ValidateDockerExecutable(ctx context.Context, docker DockerExecutable, os string) error {
	err := CheckMinimumDockerVersion(ctx, docker)
	if err != nil {
		return fmt.Errorf("failed to validate docker: %v", err)
	}
	if os == "darwin" {
		err = CheckDockerDesktopVersion(ctx, docker)
		if err != nil {
			return fmt.Errorf("failed to validate docker desktop: %v", err)
		}
	}
	CheckDockerAllocatedMemory(ctx, docker)

	return nil
}
