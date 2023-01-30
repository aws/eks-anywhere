package createcluster

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/gitops/flux"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/validations"
)

type ValidationManager struct {
	clusterSpec       *cluster.Spec
	provider          providers.Provider
	gitOpsFlux        *flux.Flux
	createValidations Validator
	dockerExec        validations.DockerExecutable
}

type Validator interface {
	PreflightValidations(ctx context.Context) []validations.Validation
}

func NewValidations(clusterSpec *cluster.Spec, provider providers.Provider, gitOpsFlux *flux.Flux, createValidations Validator, dockerExec validations.DockerExecutable) *ValidationManager {
	return &ValidationManager{
		clusterSpec:       clusterSpec,
		provider:          provider,
		gitOpsFlux:        gitOpsFlux,
		createValidations: createValidations,
		dockerExec:        dockerExec,
	}
}

func (v *ValidationManager) Validate(ctx context.Context) error {
	runner := validations.NewRunner()
	runner.Register(v.generateCreateValidations(ctx)...)
	runner.Register(v.gitOpsFlux.Validations(ctx, v.clusterSpec)...)
	err := runner.Run()

	return err
}

func (v *ValidationManager) generateCreateValidations(ctx context.Context) []validations.Validation {
	vs := []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:   "validate docker executable",
				Err:    validations.ValidateDockerExecutable(ctx, v.dockerExec, runtime.GOOS),
				Silent: true,
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:   "validate kubeconfig path",
				Err:    kubeconfig.ValidateKubeconfigPath(v.clusterSpec.Cluster.Name),
				Silent: true,
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:   "validate cluster",
				Err:    cluster.ValidateConfig(v.clusterSpec.Config),
				Silent: true,
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("validate %s Provider", v.provider.Name()),
				Err:  v.provider.SetupAndValidateCreateCluster(ctx, v.clusterSpec),
			}
		},
	}

	vs = append(vs, v.createValidations.PreflightValidations(ctx)...)

	return vs
}
