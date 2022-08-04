package workflows

import (
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/cmdvalidations"
)

type ValidateCmd struct {
	Runner        *validations.Runner
	ClusterConfig *cluster.Config
	clusterSpec   *cluster.Spec
}

func NewValidate() *ValidateCmd {
	runner := validations.NewRunner()
	return &ValidateCmd{
		Runner: runner,
	}
}

func (v *ValidateCmd) RunConfigValidations(clusterConfig *cluster.Config) error {
	v.ClusterConfig = clusterConfig
	v.Runner.Register(cmdvalidations.PackageClusterValidation(v.ClusterConfig.Cluster)...)
	err := v.Runner.Run()
	if err != nil {
		return err
	}

	v.Runner.Register(cmdvalidations.PackageKubeConfigPath(v.ClusterConfig.Cluster.Name)...)
	err = v.Runner.Run()
	if err != nil {
		return err
	}
	return nil
}
