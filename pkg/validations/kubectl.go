package validations

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

type KubectlClient interface {
	ValidateControlPlaneNodes(ctx context.Context, cluster *types.Cluster, clusterName string) error
	ValidateWorkerNodes(ctx context.Context, clusterName string, kubeconfig string) error
	ValidateNodes(ctx context.Context, kubeconfig string) error
	ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error
	ValidateEKSAClustersCRD(ctx context.Context, cluster *types.Cluster) error
	Version(ctx context.Context, cluster *types.Cluster) (*executables.VersionResponse, error)
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error)
	GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	GetEksaAWSIamConfig(ctx context.Context, awsIamConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSIamConfig, error)
	SearchEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) ([]*v1alpha1.GitOpsConfig, error)
	SearchIdentityProviderConfig(ctx context.Context, ipName string, kind string, kubeconfigFile string, namespace string) ([]*v1alpha1.VSphereDatacenterConfig, error)
}

func NewKubectl(t *testing.T) (*executables.Kubectl, context.Context, *types.Cluster, *mockexecutables.MockExecutable) {
	kubeconfigFile := "c.kubeconfig"
	cluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
	}

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)

	return executables.NewKubectl(executable), ctx, cluster, executable
}
