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

type ValidateCmd struct {
	ctx    context.Context
	Runner *validations.Runner
}

func NewValidate(ctx context.Context) *ValidateCmd {
	runner := validations.NewRunner()
	return &ValidateCmd{
		ctx:    ctx,
		Runner: runner,
	}
}

func (v *ValidateCmd) RunConfigValidations(clusterConfig *cluster.Config) error {
	v.Runner.Register(cmdvalidations.PackageClusterValidation(clusterConfig.Cluster)...)
	err := v.Runner.Run()
	if err != nil {
		return err
	}

	v.Runner.Register(cmdvalidations.PackageKubeConfigPath(clusterConfig.Cluster.Name)...)
	err = v.Runner.Run()
	if err != nil {
		return err
	}
	return nil
}

func (v *ValidateCmd) RunDockerValidations() {
	v.Runner.Register(cmdvalidations.PackageDockerValidations(v.ctx)...)
	v.Runner.Run()
}

func (v *ValidateCmd) RunSpecValidations(clusterSpec *cluster.Spec, deps *dependencies.Dependencies,
	cliConfig *config.CliConfig,
) error {
	v.Runner.Register(cmdvalidations.PackageSupportedProvider(deps.Provider)...)
	err := v.Runner.Run()
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

	v.Runner.Register(cmdvalidations.PackageCreatePreflight(v.ctx, createValidations)...)
	v.Runner.Register(cmdvalidations.PackageProviderValidations(v.ctx, clusterSpec, deps.Provider)...)
	v.Runner.Register(deps.GitOpsFlux.Validations(v.ctx, clusterSpec)...)
	err = v.Runner.Run()

	return err
}
