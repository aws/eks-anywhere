package createvalidations

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

type ValidationsKubectlClient interface {
	ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error)
	GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	GetEksaAWSIamConfig(ctx context.Context, awsIamConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSIamConfig, error)
}

func New(opts *CreateValidationOpts) *CreateValidations {
	return &CreateValidations{Opts: opts}
}

type CreateValidations struct {
	Opts *CreateValidationOpts
}

type CreateValidationOpts struct {
	Kubectl           ValidationsKubectlClient
	Spec              *cluster.Spec
	WorkloadCluster   *types.Cluster
	ManagementCluster *types.Cluster
	Provider          providers.Provider
}
