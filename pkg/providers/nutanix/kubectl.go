package nutanix

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/types"
)

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	SetEksaControllerEnvVar(ctx context.Context, envVar, envVarVal, kubeconfig string) error
}
