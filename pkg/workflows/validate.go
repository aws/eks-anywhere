package workflows

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/cmdvalidations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
)

type Validate struct {
	ctx    context.Context
	runner *validations.Runner
}

func NewValidate(ctx context.Context) *Validate {
	runner := validations.NewRunner()
	return &Validate{
		ctx:    ctx,
		runner: runner,
	}
}

func (v *Validate) RunConfigValidations(clusterConfig *cluster.Config) error {
	v.runner.Register(cmdvalidations.PackageClusterValidation(clusterConfig.Cluster)...)
	err := v.runner.Run()
	if err != nil {
		return err
	}

	v.runner.Register(cmdvalidations.PackageKubeConfigPath(clusterConfig.Cluster.Name)...)
	err = v.runner.Run()
	if err != nil {
		return err
	}
	return nil
}

func (v *Validate) RunDockerValidations() {
	v.runner.Register(cmdvalidations.PackageDockerValidations(v.ctx)...)
	v.runner.Run()
}

func (v *Validate) RunSpecValidations(clusterSpec *cluster.Spec, deps *dependencies.Dependencies,
	cliConfig *config.CliConfig,
) error {
	v.runner.Register(cmdvalidations.PackageSupportedProvider(deps.Provider)...)
	err := v.runner.Run()
	if err != nil {
		return err
	}

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		}
	} else {
		cluster = &types.Cluster{
			Name:           clusterSpec.ManagementCluster.Name,
			KubeconfigFile: clusterSpec.ManagementCluster.KubeconfigFile,
		}
	}

	validationOpts := &validations.Opts{
		Kubectl: deps.Kubectl,
		Spec:    clusterSpec,
		WorkloadCluster: &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		},
		ManagementCluster: cluster,
		Provider:          deps.Provider,
		CliConfig:         cliConfig,
	}

	createValidations := createvalidations.New(validationOpts)

	v.runner.Register(cmdvalidations.PackageCreatePreflight(v.ctx, createValidations)...)
	v.runner.Register(cmdvalidations.PackageProviderValidations(v.ctx, clusterSpec, deps.Provider)...)
	v.runner.Register(deps.GitOpsFlux.Validations(v.ctx, clusterSpec)...)
	err = v.runner.Run()

	return err
}
