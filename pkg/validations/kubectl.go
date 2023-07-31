package validations

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type KubectlClient interface {
	List(ctx context.Context, kubeconfig string, list kubernetes.ObjectList) error
	ValidateControlPlaneNodes(ctx context.Context, cluster *types.Cluster, clusterName string) error
	ValidateWorkerNodes(ctx context.Context, clusterName string, kubeconfig string) error
	ValidateNodes(ctx context.Context, kubeconfig string) error
	ValidateClustersCRD(ctx context.Context, cluster *types.Cluster) error
	ValidateEKSAClustersCRD(ctx context.Context, cluster *types.Cluster) error
	Version(ctx context.Context, cluster *types.Cluster) (*executables.VersionResponse, error)
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetBundles(ctx context.Context, kubeconfigFile, name, namespace string) (*releasev1alpha1.Bundles, error)
	GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error)
	GetEksaFluxConfig(ctx context.Context, fluxConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.FluxConfig, error)
	GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, vsphereDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	GetEksaTinkerbellDatacenterConfig(ctx context.Context, tinkerbellDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellDatacenterConfig, error)
	GetEksaTinkerbellMachineConfig(ctx context.Context, tinkerbellMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.TinkerbellMachineConfig, error)
	GetEksaAWSIamConfig(ctx context.Context, awsIamConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSIamConfig, error)
	SearchIdentityProviderConfig(ctx context.Context, ipName string, kind string, kubeconfigFile string, namespace string) ([]*v1alpha1.VSphereDatacenterConfig, error)
	GetObject(ctx context.Context, resourceType, name, namespace, kubeconfig string, obj runtime.Object) error
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
