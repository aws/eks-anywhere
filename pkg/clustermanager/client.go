package clustermanager

import (
	"context"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// KubernetesClient allows to interact with the k8s api server.
type KubernetesClient interface {
	Apply(ctx context.Context, kubeconfigPath string, obj runtime.Object, opts ...kubernetes.KubectlApplyOption) error
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error
	ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error
	WaitForDeployment(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error
	UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error
	RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error
	CreateNamespaceIfNotPresent(ctx context.Context, kubeconfig string, namespace string) error
	PauseCAPICluster(ctx context.Context, cluster, kubeconfig string) error
	ResumeCAPICluster(ctx context.Context, cluster, kubeconfig string) error
	ListObjects(ctx context.Context, resourceType, namespace, kubeconfig string, list kubernetes.ObjectList) error
	DeleteGitOpsConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error
	DeleteEKSACluster(ctx context.Context, cluster *types.Cluster, name string, namespace string) error
	DeleteAWSIamConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error
	DeleteOIDCConfig(ctx context.Context, cluster *types.Cluster, name string, namespace string) error
	DeleteCluster(ctx context.Context, cluster, clusterToDelete *types.Cluster) error
	WaitForClusterReady(ctx context.Context, cluster *types.Cluster, timeout string, clusterName string) error
	WaitForControlPlaneAvailable(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	WaitForControlPlaneReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	WaitForControlPlaneNotReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	WaitForManagedExternalEtcdReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	WaitForManagedExternalEtcdNotReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error)
	GetEksaFluxConfig(ctx context.Context, fluxConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.FluxConfig, error)
	GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error)
	GetEksaAWSIamConfig(ctx context.Context, awsIamConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSIamConfig, error)
	SaveLog(ctx context.Context, cluster *types.Cluster, deployment *types.Deployment, fileName string, writer filewriter.FileWriter) error
	GetMachines(ctx context.Context, cluster *types.Cluster, clusterName string) ([]types.Machine, error)
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, VSphereDatacenterName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	UpdateEnvironmentVariablesInNamespace(ctx context.Context, resourceType, resourceName string, envMap map[string]string, cluster *types.Cluster, namespace string) error
	GetEksaVSphereMachineConfig(ctx context.Context, VSphereDatacenterName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereMachineConfig, error)
	GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error)
	SetEksaControllerEnvVar(ctx context.Context, envVar, envVarVal, kubeconfig string) error
	DeletePackageResources(ctx context.Context, managementCluster *types.Cluster, clusterName string) error
	ValidateControlPlaneNodes(ctx context.Context, cluster *types.Cluster, clusterName string) error
	ValidateWorkerNodes(ctx context.Context, clusterName string, kubeconfigFile string) error
	CountMachineDeploymentReplicasReady(ctx context.Context, clusterName string, kubeconfigFile string) (int, int, error)
	GetBundles(ctx context.Context, kubeconfigFile, name, namespace string) (*releasev1alpha1.Bundles, error)
	GetApiServerUrl(ctx context.Context, cluster *types.Cluster) (string, error)
	KubeconfigSecretAvailable(ctx context.Context, kubeconfig string, clusterName string, namespace string) (bool, error)
	DeleteOldWorkerNodeGroup(ctx context.Context, machineDeployment *clusterv1.MachineDeployment, kubeconfig string) error
	GetKubeadmControlPlane(ctx context.Context, cluster *types.Cluster, clusterName string, opts ...executables.KubectlOpt) (*controlplanev1.KubeadmControlPlane, error)
	GetMachineDeploymentsForCluster(ctx context.Context, clusterName string, opts ...executables.KubectlOpt) ([]clusterv1.MachineDeployment, error)
	GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetEksdRelease(ctx context.Context, name, namespace, kubeconfigFile string) (*eksdv1alpha1.Release, error)
	GetConfigMap(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.ConfigMap, error)
}

// ClusterClient is an interface that has both the clusterctl client and the kubernetes retrier client.
type ClusterClient interface {
	CAPIClient
	KubernetesClient
}
