package cmdvalidations

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

func PackageDockerValidations(ctx context.Context) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: "validate docker executable",
				Err:  validateDockerExecutable(ctx),
			}
		},
	}
}

func PackageKubeConfigPath(clusterName string) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: "validate kubeconfig path",
				Err:  validateKubeConfigPath(clusterName),
			}
		},
	}
}

func PackageSupportedProvider(provider providers.Provider) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: "validate supported provider",
				Err:  validateSupportedProvider(provider),
			}
		},
	}
}

func PackageCreatePreflight(ctx context.Context, createValidations interfaces.Validator) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: "create preflight validations pass",
				Err:  createValidations.PreflightValidations(ctx),
			}
		},
	}
}

func PackageProviderValidations(ctx context.Context, clusterSpec *cluster.Spec, provider providers.Provider) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("%s Provider setup is valid", provider.Name()),
				Err:  provider.SetupAndValidateCreateCluster(ctx, clusterSpec),
			}
		},
	}
}

func PackageClusterValidation(cluster *v1alpha1.Cluster) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: "validate cluster",
				Err:  cluster.Validate(),
			}
		},
	}
}

func validateDockerExecutable(ctx context.Context) error {
	docker := executables.BuildDockerExecutable()
	err := validations.CheckMinimumDockerVersion(ctx, docker)
	if err != nil {
		return fmt.Errorf("failed to validate docker: %v", err)
	}
	if runtime.GOOS == "darwin" {
		err = validations.CheckDockerDesktopVersion(ctx, docker)
		if err != nil {
			return fmt.Errorf("failed to validate docker desktop: %v", err)
		}
	}
	validations.CheckDockerAllocatedMemory(ctx, docker)

	return nil
}

func validateKubeConfigPath(clusterName string) error {
	kubeconfigPath := kubeconfig.FromClusterName(clusterName)
	if validations.FileExistsAndIsNotEmpty(kubeconfigPath) {
		return fmt.Errorf(
			"old cluster config file exists under %s, please use a different clusterName to proceed",
			clusterName,
		)
	}

	return nil
}

func validateSupportedProvider(provider providers.Provider) error {
	if !features.IsActive(features.CloudStackProvider()) && provider.Name() == constants.CloudStackProviderName {
		return fmt.Errorf("provider cloudstack is not supported in this release")
	}

	if !features.IsActive(features.SnowProvider()) && provider.Name() == constants.SnowProviderName {
		return fmt.Errorf("provider snow is not supported in this release")
	}

	return nil
}
