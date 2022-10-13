package nutanix

import (
	"context"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/types"
)

type ProviderKubectlClient interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	SetEksaControllerEnvVar(ctx context.Context, envVar, envVarVal, kubeconfig string) error
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaNutanixDatacenterConfig(ctx context.Context, nutanixDatacenterConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.NutanixDatacenterConfig, error)
	GetEksaNutanixMachineConfig(ctx context.Context, nutanixMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.NutanixMachineConfig, error)
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*kubeadmv1beta1.KubeadmControlPlane, error)
	GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	SearchNutanixMachineConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.NutanixMachineConfig, error)
	SearchNutanixDatacenterConfig(ctx context.Context, name string, kubeconfigFile string, namespace string) ([]*v1alpha1.NutanixDatacenterConfig, error)
	DeleteEksaNutanixDatacenterConfig(ctx context.Context, nutanixDatacenterConfigName string, kubeconfigFile string, namespace string) error
	DeleteEksaNutanixMachineConfig(ctx context.Context, nutanixMachineConfigName string, kubeconfigFile string, namespace string) error
}
