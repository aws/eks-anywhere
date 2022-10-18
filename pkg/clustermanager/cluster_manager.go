package clustermanager

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	maxRetries                = 30
	defaultBackOffPeriod      = 5 * time.Second
	machineBackoff            = 1 * time.Second
	defaultMachinesMinWait    = 30 * time.Minute
	moveCAPIWait              = 15 * time.Minute
	DefaultMaxWaitPerMachine  = 10 * time.Minute
	clusterWaitStr            = "60m"
	DefaultControlPlaneWait   = 60 * time.Minute
	deploymentWaitStr         = "30m"
	controlPlaneInProgressStr = "1m"
	etcdInProgressStr         = "1m"
	DefaultEtcdWait           = 60 * time.Minute
)

var eksaClusterResourceType = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)

type ClusterManager struct {
	*Upgrader
	clusterClient           *retrierClient
	writer                  filewriter.FileWriter
	networking              Networking
	diagnosticsFactory      diagnostics.DiagnosticBundleFactory
	Retrier                 *retrier.Retrier
	machineMaxWait          time.Duration
	machineBackoff          time.Duration
	machinesMinWait         time.Duration
	awsIamAuth              AwsIamAuth
	controlPlaneWaitTimeout time.Duration
	externalEtcdWaitTimeout time.Duration
}

type ClusterClient interface {
	MoveManagement(ctx context.Context, org, target *types.Cluster) error
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error
	ApplyKubeSpecFromBytesForce(ctx context.Context, cluster *types.Cluster, data []byte) error
	WaitForClusterReady(ctx context.Context, cluster *types.Cluster, timeout string, clusterName string) error
	WaitForControlPlaneReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	WaitForControlPlaneNotReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	WaitForManagedExternalEtcdReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	WaitForManagedExternalEtcdNotReady(ctx context.Context, cluster *types.Cluster, timeout string, newClusterName string) error
	GetWorkloadKubeconfig(ctx context.Context, clusterName string, cluster *types.Cluster) ([]byte, error)
	GetEksaGitOpsConfig(ctx context.Context, gitOpsConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.GitOpsConfig, error)
	GetEksaFluxConfig(ctx context.Context, fluxConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.FluxConfig, error)
	GetEksaOIDCConfig(ctx context.Context, oidcConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.OIDCConfig, error)
	GetEksaAWSIamConfig(ctx context.Context, awsIamConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.AWSIamConfig, error)
	DeleteCluster(ctx context.Context, managementCluster, clusterToDelete *types.Cluster) error
	DeleteGitOpsConfig(ctx context.Context, managementCluster *types.Cluster, gitOpsName, namespace string) error
	DeleteOIDCConfig(ctx context.Context, managementCluster *types.Cluster, oidcConfigName, oidcConfigNamespace string) error
	DeleteAWSIamConfig(ctx context.Context, managementCluster *types.Cluster, awsIamConfigName, awsIamConfigNamespace string) error
	DeleteEKSACluster(ctx context.Context, managementCluster *types.Cluster, eksaClusterName, eksaClusterNamespace string) error
	DeletePackageResources(ctx context.Context, managementCluster *types.Cluster, clusterName string) error
	InitInfrastructure(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error
	WaitForDeployment(ctx context.Context, cluster *types.Cluster, timeout string, condition string, target string, namespace string) error
	SaveLog(ctx context.Context, cluster *types.Cluster, deployment *types.Deployment, fileName string, writer filewriter.FileWriter) error
	GetMachines(ctx context.Context, cluster *types.Cluster, clusterName string) ([]types.Machine, error)
	GetClusters(ctx context.Context, cluster *types.Cluster) ([]types.CAPICluster, error)
	GetEksaCluster(ctx context.Context, cluster *types.Cluster, clusterName string) (*v1alpha1.Cluster, error)
	GetEksaVSphereDatacenterConfig(ctx context.Context, VSphereDatacenterName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereDatacenterConfig, error)
	UpdateEnvironmentVariablesInNamespace(ctx context.Context, resourceType, resourceName string, envMap map[string]string, cluster *types.Cluster, namespace string) error
	UpdateAnnotationInNamespace(ctx context.Context, resourceType, objectName string, annotations map[string]string, cluster *types.Cluster, namespace string) error
	RemoveAnnotationInNamespace(ctx context.Context, resourceType, objectName, key string, cluster *types.Cluster, namespace string) error
	GetEksaVSphereMachineConfig(ctx context.Context, VSphereDatacenterName string, kubeconfigFile string, namespace string) (*v1alpha1.VSphereMachineConfig, error)
	GetEksaCloudStackMachineConfig(ctx context.Context, cloudstackMachineConfigName string, kubeconfigFile string, namespace string) (*v1alpha1.CloudStackMachineConfig, error)
	SetEksaControllerEnvVar(ctx context.Context, envVar, envVarVal, kubeconfig string) error
	CreateNamespaceIfNotPresent(ctx context.Context, kubeconfig string, namespace string) error
	ValidateControlPlaneNodes(ctx context.Context, cluster *types.Cluster, clusterName string) error
	ValidateWorkerNodes(ctx context.Context, clusterName string, kubeconfigFile string) error
	CountMachineDeploymentReplicasReady(ctx context.Context, clusterName string, kubeconfigFile string) (int, int, error)
	GetBundles(ctx context.Context, kubeconfigFile, name, namespace string) (*releasev1alpha1.Bundles, error)
	GetApiServerUrl(ctx context.Context, cluster *types.Cluster) (string, error)
	GetClusterCATlsCert(ctx context.Context, clusterName string, cluster *types.Cluster, namespace string) ([]byte, error)
	KubeconfigSecretAvailable(ctx context.Context, kubeconfig string, clusterName string, namespace string) (bool, error)
	DeleteOldWorkerNodeGroup(ctx context.Context, machineDeployment *clusterv1.MachineDeployment, kubeconfig string) error
	GetMachineDeployment(ctx context.Context, workerNodeGroupName string, opts ...executables.KubectlOpt) (*clusterv1.MachineDeployment, error)
	GetEksdRelease(ctx context.Context, name, namespace, kubeconfigFile string) (*eksdv1alpha1.Release, error)
	ListObjects(ctx context.Context, resourceType, namespace, kubeconfig string, list kubernetes.ObjectList) error
}

type Networking interface {
	GenerateManifest(ctx context.Context, clusterSpec *cluster.Spec, namespaces []string) ([]byte, error)
	Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, namespaces []string) (*types.ChangeDiff, error)
	RunPostControlPlaneUpgradeSetup(ctx context.Context, cluster *types.Cluster) error
}

type AwsIamAuth interface {
	GenerateManifest(clusterSpec *cluster.Spec) ([]byte, error)
	GenerateManifestForUpgrade(clusterSpec *cluster.Spec) ([]byte, error)
	GenerateCertKeyPairSecret(managementClusterName string) ([]byte, error)
	GenerateAwsIamAuthKubeconfig(clusterSpec *cluster.Spec, serverUrl, tlsCert string) ([]byte, error)
}

type ClusterManagerOpt func(*ClusterManager)

func New(clusterClient ClusterClient, networking Networking, writer filewriter.FileWriter, diagnosticBundleFactory diagnostics.DiagnosticBundleFactory, awsIamAuth AwsIamAuth, opts ...ClusterManagerOpt) *ClusterManager {
	retrier := retrier.NewWithMaxRetries(maxRetries, defaultBackOffPeriod)
	retrierClient := NewRetrierClient(NewClient(clusterClient), retrier)
	c := &ClusterManager{
		Upgrader:                NewUpgrader(retrierClient),
		clusterClient:           retrierClient,
		writer:                  writer,
		networking:              networking,
		Retrier:                 retrier,
		diagnosticsFactory:      diagnosticBundleFactory,
		machineMaxWait:          DefaultMaxWaitPerMachine,
		machineBackoff:          machineBackoff,
		machinesMinWait:         defaultMachinesMinWait,
		awsIamAuth:              awsIamAuth,
		controlPlaneWaitTimeout: DefaultControlPlaneWait,
		externalEtcdWaitTimeout: DefaultEtcdWait,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func WithControlPlaneWaitTimeout(timeout time.Duration) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.controlPlaneWaitTimeout = timeout
	}
}

func WithExternalEtcdWaitTimeout(timeout time.Duration) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.externalEtcdWaitTimeout = timeout
	}
}

func WithMachineBackoff(machineBackoff time.Duration) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.machineBackoff = machineBackoff
	}
}

func WithMachineMaxWait(machineMaxWait time.Duration) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.machineMaxWait = machineMaxWait
	}
}

func WithMachineMinWait(machineMinWait time.Duration) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.machinesMinWait = machineMinWait
	}
}

func WithRetrier(retrier *retrier.Retrier) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.clusterClient.Retrier = retrier
		c.Retrier = retrier
	}
}

func (c *ClusterManager) MoveCAPI(ctx context.Context, from, to *types.Cluster, clusterName string, clusterSpec *cluster.Spec, checkers ...types.NodeReadyChecker) error {
	logger.V(3).Info("Waiting for management machines to be ready before move")
	labels := []string{clusterv1.MachineControlPlaneLabelName, clusterv1.MachineDeploymentLabelName}
	if err := c.waitForNodesReady(ctx, from, clusterName, labels, checkers...); err != nil {
		return err
	}

	logger.V(3).Info("Waiting for all clusters to be ready before move")
	if err := c.waitForAllClustersReady(ctx, from, clusterWaitStr); err != nil {
		return err
	}

	err := c.clusterClient.MoveManagement(ctx, from, to)
	if err != nil {
		return fmt.Errorf("moving CAPI management from source to target: %v", err)
	}

	logger.V(3).Info("Waiting for control planes to be ready after move")
	err = c.waitForAllControlPlanes(ctx, to, moveCAPIWait)
	if err != nil {
		return err
	}

	logger.V(3).Info("Waiting for workload cluster control plane replicas to be ready after move")
	err = c.waitForControlPlaneReplicasReady(ctx, to, clusterSpec)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster control plane replicas to be ready: %v", err)
	}

	logger.V(3).Info("Waiting for workload cluster machine deployment replicas to be ready after move")
	err = c.waitForMachineDeploymentReplicasReady(ctx, to, clusterSpec)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster machinedeployment replicas to be ready: %v", err)
	}

	logger.V(3).Info("Waiting for machines to be ready after move")
	if err = c.waitForNodesReady(ctx, to, clusterName, labels, checkers...); err != nil {
		return err
	}

	return nil
}

func (c *ClusterManager) writeCAPISpecFile(clusterName string, content []byte) error {
	fileName := fmt.Sprintf("%s-eks-a-cluster.yaml", clusterName)
	if _, err := c.writer.Write(fileName, content); err != nil {
		return fmt.Errorf("writing capi spec file: %v", err)
	}
	return nil
}

// CreateWorkloadCluster creates a workload cluster in the provider that the customer has specified.
// It applied the kubernetes manifest file on the management cluster, waits for the control plane to be ready,
// and then generates the kubeconfig for the cluster.
// It returns a struct of type Cluster containing the name and the kubeconfig of the cluster.
func (c *ClusterManager) CreateWorkloadCluster(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) (*types.Cluster, error) {
	workloadCluster := &types.Cluster{
		Name:               clusterSpec.Cluster.Name,
		ExistingManagement: managementCluster.ExistingManagement,
	}

	cpContent, mdContent, err := provider.GenerateCAPISpecForCreate(ctx, managementCluster, clusterSpec)
	if err != nil {
		return nil, fmt.Errorf("generating capi spec: %v", err)
	}

	content := templater.AppendYamlResources(cpContent, mdContent)

	if err = c.writeCAPISpecFile(clusterSpec.Cluster.Name, content); err != nil {
		return nil, err
	}

	err = c.clusterClient.ApplyKubeSpecFromBytesWithNamespace(ctx, managementCluster, content, constants.EksaSystemNamespace)
	if err != nil {
		return nil, fmt.Errorf("applying capi spec: %v", err)
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		logger.V(3).Info("Waiting for external etcd to be ready", "cluster", workloadCluster.Name)
		err = c.clusterClient.WaitForManagedExternalEtcdReady(ctx, managementCluster, c.externalEtcdWaitTimeout.String(), workloadCluster.Name)
		if err != nil {
			return nil, fmt.Errorf("waiting for external etcd for workload cluster to be ready: %v", err)
		}
		logger.V(3).Info("External etcd is ready")
		// the condition external etcd ready if true indicates that all etcd machines are ready and the etcd cluster is ready to accept requests
	}

	logger.V(3).Info("Waiting for control plane to be ready")
	err = c.clusterClient.WaitForControlPlaneReady(ctx, managementCluster, c.controlPlaneWaitTimeout.String(), workloadCluster.Name)
	if err != nil {
		return nil, fmt.Errorf("waiting for workload cluster control plane to be ready: %v", err)
	}

	logger.V(3).Info("Waiting for workload kubeconfig generation", "cluster", workloadCluster.Name)
	err = c.Retrier.Retry(
		func() error {
			workloadCluster.KubeconfigFile, err = c.generateWorkloadKubeconfig(ctx, workloadCluster.Name, managementCluster, provider)
			return err
		},
	)
	if err != nil {
		return nil, fmt.Errorf("waiting for workload kubeconfig: %v", err)
	}

	return workloadCluster, nil
}

func (c *ClusterManager) generateWorkloadKubeconfig(ctx context.Context, clusterName string, cluster *types.Cluster, provider providers.Provider) (string, error) {
	fileName := fmt.Sprintf("%s-eks-a-cluster.kubeconfig", clusterName)
	kubeconfig, err := c.clusterClient.GetWorkloadKubeconfig(ctx, clusterName, cluster)
	if err != nil {
		return "", fmt.Errorf("getting workload kubeconfig: %v", err)
	}
	if err := provider.UpdateKubeConfig(&kubeconfig, clusterName); err != nil {
		return "", err
	}

	writtenFile, err := c.writer.Write(fileName, kubeconfig, filewriter.PersistentFile, filewriter.Permission0600)
	if err != nil {
		return "", fmt.Errorf("writing workload kubeconfig: %v", err)
	}
	return writtenFile, nil
}

func (c *ClusterManager) RunPostCreateWorkloadCluster(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.V(3).Info("Waiting for controlplane and worker machines to be ready")
	labels := []string{clusterv1.MachineControlPlaneLabelName, clusterv1.MachineDeploymentLabelName}
	return c.waitForNodesReady(ctx, managementCluster, workloadCluster.Name, labels, types.WithNodeRef())
}

func (c *ClusterManager) DeleteCluster(ctx context.Context, managementCluster, clusterToDelete *types.Cluster, provider providers.Provider, clusterSpec *cluster.Spec) error {
	return c.Retrier.Retry(
		func() error {
			if clusterSpec.Cluster.IsManaged() {
				if err := c.PauseEKSAControllerReconcile(ctx, clusterToDelete, clusterSpec, provider); err != nil {
					return err
				}

				if clusterSpec.GitOpsConfig != nil {
					if err := c.DeleteGitOpsConfig(ctx, managementCluster, clusterSpec.GitOpsConfig.Name, clusterSpec.GitOpsConfig.Namespace); err != nil {
						return err
					}
				}
				if clusterSpec.OIDCConfig != nil {
					if err := c.DeleteOIDCConfig(ctx, managementCluster, clusterSpec.OIDCConfig.Name, clusterSpec.OIDCConfig.Namespace); err != nil {
						return err
					}
				}

				if clusterSpec.AWSIamConfig != nil {
					if err := c.DeleteAWSIamConfig(ctx, managementCluster, clusterSpec.AWSIamConfig.Name, clusterSpec.AWSIamConfig.Namespace); err != nil {
						return err
					}
				}

				if err := provider.DeleteResources(ctx, clusterSpec); err != nil {
					return err
				}

				if err := c.DeleteEKSACluster(ctx, managementCluster, clusterSpec.Cluster.Name, clusterSpec.Cluster.Namespace); err != nil {
					return err
				}
			}

			if err := c.clusterClient.DeleteCluster(ctx, managementCluster, clusterToDelete); err != nil {
				return err
			}

			return provider.PostClusterDeleteValidate(ctx, managementCluster)
		},
	)
}

func (c *ClusterManager) UpgradeCluster(ctx context.Context, managementCluster, workloadCluster *types.Cluster, newClusterSpec *cluster.Spec, provider providers.Provider) error {
	eksaMgmtCluster := workloadCluster
	if managementCluster != nil && managementCluster.ExistingManagement {
		eksaMgmtCluster = managementCluster
	}

	currentSpec, err := c.GetCurrentClusterSpec(ctx, eksaMgmtCluster, newClusterSpec.Cluster.Name)
	if err != nil {
		return fmt.Errorf("getting current cluster spec: %v", err)
	}

	cpContent, mdContent, err := provider.GenerateCAPISpecForUpgrade(ctx, managementCluster, eksaMgmtCluster, currentSpec, newClusterSpec)
	if err != nil {
		return fmt.Errorf("generating capi spec: %v", err)
	}

	if err = c.writeCAPISpecFile(newClusterSpec.Cluster.Name, templater.AppendYamlResources(cpContent, mdContent)); err != nil {
		return err
	}
	err = c.clusterClient.ApplyKubeSpecFromBytesWithNamespace(ctx, managementCluster, cpContent, constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("applying capi control plane spec: %v", err)
	}

	var externalEtcdTopology bool
	if newClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		logger.V(3).Info("Waiting for external etcd upgrade to be in progress")
		err = c.clusterClient.WaitForManagedExternalEtcdNotReady(ctx, managementCluster, etcdInProgressStr, newClusterSpec.Cluster.Name)
		if err != nil {
			if !strings.Contains(fmt.Sprint(err), "timed out waiting for the condition on clusters") {
				return fmt.Errorf("error waiting for external etcd upgrade not ready: %v", err)
			} else {
				logger.V(3).Info("Timed out while waiting for external etcd to be in progress, likely caused by no external etcd upgrade")
			}
		}

		logger.V(3).Info("Waiting for external etcd to be ready after upgrade")
		if err = c.clusterClient.WaitForManagedExternalEtcdReady(ctx, managementCluster, c.externalEtcdWaitTimeout.String(), newClusterSpec.Cluster.Name); err != nil {
			if err := c.clusterClient.RemoveAnnotationInNamespace(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", newClusterSpec.Cluster.Name),
				etcdv1.UpgradeInProgressAnnotation,
				managementCluster,
				constants.EksaSystemNamespace); err != nil {
				return fmt.Errorf("removing annotation: %v", err)
			}
			return fmt.Errorf("waiting for external etcd for workload cluster to be ready: %v", err)
		}
		externalEtcdTopology = true
		logger.V(3).Info("External etcd is ready")
	}

	logger.V(3).Info("Waiting for control plane upgrade to be in progress")
	err = c.clusterClient.WaitForControlPlaneNotReady(ctx, managementCluster, controlPlaneInProgressStr, newClusterSpec.Cluster.Name)
	if err != nil {
		if !strings.Contains(fmt.Sprint(err), "timed out waiting for the condition on clusters") {
			return fmt.Errorf("error waiting for control plane not ready: %v", err)
		} else {
			logger.V(3).Info("Timed out while waiting for control plane to be in progress, likely caused by no control plane upgrade")
		}
	}
	logger.V(3).Info("Run post control plane upgrade operations")
	err = provider.RunPostControlPlaneUpgrade(ctx, currentSpec, newClusterSpec, workloadCluster, managementCluster)
	if err != nil {
		return fmt.Errorf("running post control plane upgrade operations: %v", err)
	}

	logger.V(3).Info("Waiting for control plane to be ready")
	err = c.clusterClient.WaitForControlPlaneReady(ctx, managementCluster, c.controlPlaneWaitTimeout.String(), newClusterSpec.Cluster.Name)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster control plane to be ready: %v", err)
	}

	logger.V(3).Info("Waiting for control plane machines to be ready")
	if err = c.waitForNodesReady(ctx, managementCluster, newClusterSpec.Cluster.Name, []string{clusterv1.MachineControlPlaneLabelName}, types.WithNodeRef(), types.WithNodeHealthy()); err != nil {
		return err
	}

	logger.V(3).Info("Waiting for control plane to be ready after upgrade")
	err = c.clusterClient.WaitForControlPlaneReady(ctx, managementCluster, c.controlPlaneWaitTimeout.String(), newClusterSpec.Cluster.Name)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster control plane to be ready: %v", err)
	}

	logger.V(3).Info("Running CNI post control plane upgrade operations")
	if err = c.networking.RunPostControlPlaneUpgradeSetup(ctx, workloadCluster); err != nil {
		return fmt.Errorf("running CNI post control plane upgrade operations: %v", err)
	}

	logger.V(3).Info("Waiting for workload cluster control plane replicas to be ready after upgrade")
	err = c.waitForControlPlaneReplicasReady(ctx, managementCluster, newClusterSpec)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster control plane replicas to be ready: %v", err)
	}

	err = c.clusterClient.ApplyKubeSpecFromBytesWithNamespace(ctx, managementCluster, mdContent, constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("applying capi machine deployment spec: %v", err)
	}

	if err = c.removeOldWorkerNodeGroups(ctx, managementCluster, provider, currentSpec, newClusterSpec); err != nil {
		return fmt.Errorf("removing old worker node groups: %v", err)
	}

	logger.V(3).Info("Waiting for workload cluster machine deployment replicas to be ready after upgrade")
	err = c.waitForMachineDeploymentReplicasReady(ctx, managementCluster, newClusterSpec)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster machinedeployment replicas to be ready: %v", err)
	}

	logger.V(3).Info("Waiting for machine deployment machines to be ready")
	if err = c.waitForNodesReady(ctx, managementCluster, newClusterSpec.Cluster.Name, []string{clusterv1.MachineDeploymentLabelName}, types.WithNodeRef(), types.WithNodeHealthy()); err != nil {
		return err
	}

	logger.V(3).Info("Waiting for workload cluster capi components to be ready after upgrade")
	err = c.waitForCAPI(ctx, eksaMgmtCluster, provider, externalEtcdTopology)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster capi components to be ready: %v", err)
	}

	if newClusterSpec.AWSIamConfig != nil {
		logger.V(3).Info("Run aws-iam-authenticator upgrade operations")
		if err = c.generateAndApplyAwsIamAuthForUpgrade(ctx, workloadCluster, newClusterSpec); err != nil {
			return fmt.Errorf("running aws-iam-authenticator upgrade operations: %v", err)
		}
	}

	if err = c.InstallStorageClass(ctx, workloadCluster, provider); err != nil {
		return fmt.Errorf("installing storage class during upgrade: %v", err)
	}

	return nil
}

func (c *ClusterManager) EKSAClusterSpecChanged(ctx context.Context, cluster *types.Cluster, newClusterSpec *cluster.Spec) (bool, error) {
	cc, err := c.clusterClient.GetEksaCluster(ctx, cluster, newClusterSpec.Cluster.Name)
	if err != nil {
		return false, err
	}

	if !cc.Equal(newClusterSpec.Cluster) {
		logger.V(3).Info("Existing cluster and new cluster spec differ")
		return true, nil
	}

	currentClusterSpec, err := c.buildSpecForCluster(ctx, cluster, cc)
	if err != nil {
		return false, err
	}

	if currentClusterSpec.VersionsBundle.EksD.Name != newClusterSpec.VersionsBundle.EksD.Name {
		logger.V(3).Info("New eks-d release detected")
		return true, nil
	}

	if newClusterSpec.OIDCConfig != nil && currentClusterSpec.OIDCConfig != nil {
		if !newClusterSpec.OIDCConfig.Spec.Equal(&currentClusterSpec.OIDCConfig.Spec) {
			logger.V(3).Info("OIDC config changes detected")
			return true, nil
		}
	}

	if newClusterSpec.AWSIamConfig != nil && currentClusterSpec.AWSIamConfig != nil {
		if !reflect.DeepEqual(newClusterSpec.AWSIamConfig.Spec.MapRoles, currentClusterSpec.AWSIamConfig.Spec.MapRoles) ||
			!reflect.DeepEqual(newClusterSpec.AWSIamConfig.Spec.MapUsers, currentClusterSpec.AWSIamConfig.Spec.MapUsers) {
			logger.V(3).Info("AWSIamConfig changes detected")
			return true, nil
		}
	}

	logger.V(3).Info("Clusters are the same")
	return false, nil
}

func (c *ClusterManager) InstallCAPI(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error {
	err := c.clusterClient.InitInfrastructure(ctx, clusterSpec, cluster, provider)
	if err != nil {
		return fmt.Errorf("initializing capi resources in cluster: %v", err)
	}

	return c.waitForCAPI(ctx, cluster, provider, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil)
}

func (c *ClusterManager) waitForCAPI(ctx context.Context, cluster *types.Cluster, provider providers.Provider, externalEtcdTopology bool) error {
	err := c.clusterClient.waitForDeployments(ctx, internal.CAPIDeployments, cluster)
	if err != nil {
		return err
	}

	if externalEtcdTopology {
		err := c.clusterClient.waitForDeployments(ctx, internal.ExternalEtcdDeployments, cluster)
		if err != nil {
			return err
		}
	}

	err = c.clusterClient.waitForDeployments(ctx, provider.GetDeployments(), cluster)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClusterManager) InstallNetworking(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error {
	providerNamespaces := getProviderNamespaces(provider.GetDeployments())
	networkingManifestContent, err := c.networking.GenerateManifest(ctx, clusterSpec, providerNamespaces)
	if err != nil {
		return fmt.Errorf("generating networking manifest: %v", err)
	}
	if err = c.clusterClient.ApplyKubeSpecFromBytes(ctx, cluster, networkingManifestContent); err != nil {
		return fmt.Errorf("applying networking manifest spec: %v", err)
	}
	return nil
}

func (c *ClusterManager) UpgradeNetworking(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, provider providers.Provider) (*types.ChangeDiff, error) {
	providerNamespaces := getProviderNamespaces(provider.GetDeployments())
	return c.networking.Upgrade(ctx, cluster, currentSpec, newSpec, providerNamespaces)
}

func getProviderNamespaces(providerDeployments map[string][]string) []string {
	namespaces := make([]string, 0, len(providerDeployments))
	for namespace := range providerDeployments {
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

func (c *ClusterManager) InstallStorageClass(ctx context.Context, cluster *types.Cluster, provider providers.Provider) error {
	storageClass := provider.GenerateStorageClass()
	if storageClass == nil {
		return nil
	}

	logger.Info("Installing storage class on cluster")
	err := c.clusterClient.ApplyKubeSpecFromBytes(ctx, cluster, storageClass)
	if err != nil {
		return fmt.Errorf("applying storage class manifest: %v", err)
	}
	return nil
}

func (c *ClusterManager) InstallMachineHealthChecks(ctx context.Context, clusterSpec *cluster.Spec, workloadCluster *types.Cluster) error {
	mhc, err := templater.ObjectsToYaml(clusterapi.MachineHealthCheckObjects(clusterSpec)...)
	if err != nil {
		return err
	}

	err = c.clusterClient.ApplyKubeSpecFromBytes(ctx, workloadCluster, mhc)
	if err != nil {
		return fmt.Errorf("applying machine health checks: %v", err)
	}
	return nil
}

// InstallAwsIamAuth applies the aws-iam-authenticator manifest based on cluster spec inputs.
// Generates a kubeconfig for interacting with the cluster with aws-iam-authenticator client.
func (c *ClusterManager) InstallAwsIamAuth(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	awsIamAuthManifest, err := c.awsIamAuth.GenerateManifest(clusterSpec)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}
	err = c.clusterClient.ApplyKubeSpecFromBytes(ctx, workloadCluster, awsIamAuthManifest)
	if err != nil {
		return fmt.Errorf("applying aws-iam-authenticator manifest: %v", err)
	}
	err = c.generateAwsIamAuthKubeconfig(ctx, managementCluster, workloadCluster, clusterSpec)
	if err != nil {
		return err
	}
	return nil
}

func (c *ClusterManager) CreateAwsIamAuthCaSecret(ctx context.Context, bootstrapCluster *types.Cluster, workloadClusterName string) error {
	awsIamAuthCaSecret, err := c.awsIamAuth.GenerateCertKeyPairSecret(workloadClusterName)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator ca secret: %v", err)
	}
	err = c.clusterClient.ApplyKubeSpecFromBytes(ctx, bootstrapCluster, awsIamAuthCaSecret)
	if err != nil {
		return fmt.Errorf("applying aws-iam-authenticator ca secret: %v", err)
	}
	return nil
}

// generateAndApplyAwsIamAuthForUpgrade generates the aws-iam-authenticator manifest based on cluster spec and applies on the cluster.
func (c *ClusterManager) generateAndApplyAwsIamAuthForUpgrade(ctx context.Context, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	awsIamAuthManifest, err := c.awsIamAuth.GenerateManifestForUpgrade(clusterSpec)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}
	err = c.clusterClient.ApplyKubeSpecFromBytes(ctx, workloadCluster, awsIamAuthManifest)
	if err != nil {
		return fmt.Errorf("applying aws-iam-authenticator manifest: %v", err)
	}
	return nil
}

func (c *ClusterManager) generateAwsIamAuthKubeconfig(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	fileName := fmt.Sprintf("%s-aws.kubeconfig", workloadCluster.Name)
	serverUrl, err := c.clusterClient.GetApiServerUrl(ctx, workloadCluster)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}
	tlsCert, err := c.clusterClient.GetClusterCATlsCert(ctx, workloadCluster.Name, managementCluster, constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}
	awsIamAuthKubeconfigContent, err := c.awsIamAuth.GenerateAwsIamAuthKubeconfig(clusterSpec, serverUrl, string(tlsCert))
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}
	writtenFile, err := c.writer.Write(fileName, awsIamAuthKubeconfigContent, filewriter.PersistentFile, filewriter.Permission0600)
	if err != nil {
		return fmt.Errorf("writing aws-iam-authenticator kubeconfig to %s: %v", writtenFile, err)
	}
	logger.V(3).Info("Generated aws-iam-authenticator kubeconfig", "kubeconfig", writtenFile)
	return nil
}

func (c *ClusterManager) SaveLogsManagementCluster(ctx context.Context, spec *cluster.Spec, cluster *types.Cluster) error {
	if cluster == nil {
		return nil
	}

	if cluster.KubeconfigFile == "" {
		return nil
	}

	bundle, err := c.diagnosticsFactory.DiagnosticBundleManagementCluster(spec, cluster.KubeconfigFile)
	if err != nil {
		logger.V(5).Info("Error generating support bundle for management cluster", "error", err)
		return nil
	}
	return collectDiagnosticBundle(ctx, bundle)
}

func (c *ClusterManager) SaveLogsWorkloadCluster(ctx context.Context, provider providers.Provider, spec *cluster.Spec, cluster *types.Cluster) error {
	if cluster == nil {
		return nil
	}

	if cluster.KubeconfigFile == "" {
		return nil
	}

	bundle, err := c.diagnosticsFactory.DiagnosticBundleWorkloadCluster(spec, provider, cluster.KubeconfigFile)
	if err != nil {
		logger.V(5).Info("Error generating support bundle for workload cluster", "error", err)
		return nil
	}

	return collectDiagnosticBundle(ctx, bundle)
}

func collectDiagnosticBundle(ctx context.Context, bundle diagnostics.DiagnosticBundle) error {
	var sinceTimeValue *time.Time
	threeHours := "3h"
	sinceTimeValue, err := diagnostics.ParseTimeFromDuration(threeHours)
	if err != nil {
		logger.V(5).Info("Error parsing time options for support bundle generation", "error", err)
		return nil
	}

	err = bundle.CollectAndAnalyze(ctx, sinceTimeValue)
	if err != nil {
		logger.V(5).Info("Error collecting and saving logs", "error", err)
	}
	return nil
}

func (c *ClusterManager) waitForControlPlaneReplicasReady(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	isCpReady := func() error {
		return c.clusterClient.ValidateControlPlaneNodes(ctx, managementCluster, clusterSpec.Cluster.Name)
	}

	err := isCpReady()
	if err == nil {
		return nil
	}

	timeout := time.Duration(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count) * c.machineMaxWait
	if timeout <= c.machinesMinWait {
		timeout = c.machinesMinWait
	}

	r := retrier.New(timeout)
	if err := r.Retry(isCpReady); err != nil {
		return fmt.Errorf("retries exhausted waiting for controlplane replicas to be ready: %v", err)
	}
	return nil
}

func (c *ClusterManager) waitForMachineDeploymentReplicasReady(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	ready, total := 0, 0
	policy := func(_ int, _ error) (bool, time.Duration) {
		return true, c.machineBackoff * time.Duration(total-ready)
	}

	var machineDeploymentReplicasCount int
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		machineDeploymentReplicasCount += *workerNodeGroupConfiguration.Count
	}

	areMdReplicasReady := func() error {
		var err error
		ready, total, err = c.clusterClient.CountMachineDeploymentReplicasReady(ctx, clusterSpec.Cluster.Name, managementCluster.KubeconfigFile)
		if err != nil {
			return err
		}
		if ready != total {
			return fmt.Errorf("%d machine deployment replicas are not ready", total-ready)
		}
		return nil
	}

	timeout := time.Duration(machineDeploymentReplicasCount) * c.machineMaxWait
	if timeout <= c.machinesMinWait {
		timeout = c.machinesMinWait
	}

	r := retrier.New(timeout, retrier.WithRetryPolicy(policy))
	if err := r.Retry(areMdReplicasReady); err != nil {
		return fmt.Errorf("retries exhausted waiting for machinedeployment replicas to be ready: %v", err)
	}
	return nil
}

func (c *ClusterManager) waitForNodesReady(ctx context.Context, managementCluster *types.Cluster, clusterName string, labels []string, checkers ...types.NodeReadyChecker) error {
	readyNodes, totalNodes := 0, 0
	policy := func(_ int, _ error) (bool, time.Duration) {
		return true, c.machineBackoff * time.Duration(totalNodes-readyNodes)
	}

	areNodesReady := func() error {
		var err error
		readyNodes, totalNodes, err = c.countNodesReady(ctx, managementCluster, clusterName, labels, checkers...)
		if err != nil {
			return err
		}

		if readyNodes != totalNodes {
			logger.V(4).Info("Nodes are not ready yet", "total", totalNodes, "ready", readyNodes, "cluster name", clusterName)
			return errors.New("nodes are not ready yet")
		}

		logger.V(4).Info("Nodes ready", "total", totalNodes)
		return nil
	}

	err := areNodesReady()
	if err == nil {
		return nil
	}

	timeout := time.Duration(totalNodes) * c.machineMaxWait
	if timeout <= c.machinesMinWait {
		timeout = c.machinesMinWait
	}

	r := retrier.New(timeout, retrier.WithRetryPolicy(policy))
	if err := r.Retry(areNodesReady); err != nil {
		return fmt.Errorf("retries exhausted waiting for machines to be ready: %v", err)
	}

	return nil
}

func (c *ClusterManager) countNodesReady(ctx context.Context, managementCluster *types.Cluster, clusterName string, labels []string, checkers ...types.NodeReadyChecker) (ready, total int, err error) {
	machines, err := c.clusterClient.GetMachines(ctx, managementCluster, clusterName)
	if err != nil {
		return 0, 0, fmt.Errorf("getting machines resources from management cluster: %v", err)
	}

	for _, m := range machines {
		// Extracted from cluster-api: NodeRef is considered a better signal than InfrastructureReady,
		// because it ensures the node in the workload cluster is up and running.
		if !m.HasAnyLabel(labels) {
			continue
		}

		total += 1

		passed := true
		for _, checker := range checkers {
			if !checker(m.Status) {
				passed = false
				break
			}
		}
		if passed {
			ready += 1
		}
	}
	return ready, total, nil
}

func (c *ClusterManager) waitForAllControlPlanes(ctx context.Context, cluster *types.Cluster, waitForCluster time.Duration) error {
	clusters, err := c.clusterClient.GetClusters(ctx, cluster)
	if err != nil {
		return fmt.Errorf("getting clusters: %v", err)
	}

	for _, clu := range clusters {
		err = c.clusterClient.WaitForControlPlaneReady(ctx, cluster, waitForCluster.String(), clu.Metadata.Name)
		if err != nil {
			return fmt.Errorf("waiting for workload cluster control plane for cluster %s to be ready: %v", clu.Metadata.Name, err)
		}
	}

	return nil
}

func (c *ClusterManager) waitForAllClustersReady(ctx context.Context, cluster *types.Cluster, waitStr string) error {
	clusters, err := c.clusterClient.GetClusters(ctx, cluster)
	if err != nil {
		return fmt.Errorf("getting clusters: %v", err)
	}

	for _, clu := range clusters {
		err = c.clusterClient.WaitForClusterReady(ctx, cluster, waitStr, clu.Metadata.Name)
		if err != nil {
			return fmt.Errorf("waiting for cluster %s to be ready: %v", clu.Metadata.Name, err)
		}
	}

	return nil
}

func machineDeploymentsToDelete(currentSpec, newSpec *cluster.Spec) []string {
	nodeGroupsToDelete := cluster.NodeGroupsToDelete(currentSpec, newSpec)
	machineDeployments := make([]string, 0, len(nodeGroupsToDelete))
	for _, group := range nodeGroupsToDelete {
		mdName := clusterapi.MachineDeploymentName(newSpec, group)
		machineDeployments = append(machineDeployments, mdName)
	}
	return machineDeployments
}

func (c *ClusterManager) removeOldWorkerNodeGroups(ctx context.Context, workloadCluster *types.Cluster, provider providers.Provider, currentSpec, newSpec *cluster.Spec) error {
	machineDeployments := machineDeploymentsToDelete(currentSpec, newSpec)
	for _, machineDeploymentName := range machineDeployments {
		machineDeployment, err := c.clusterClient.GetMachineDeployment(ctx, machineDeploymentName, executables.WithKubeconfig(workloadCluster.KubeconfigFile), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return fmt.Errorf("getting machine deployment to remove: %v", err)
		}
		if err := c.clusterClient.DeleteOldWorkerNodeGroup(ctx, machineDeployment, workloadCluster.KubeconfigFile); err != nil {
			return fmt.Errorf("removing old worker nodes from cluster: %v", err)
		}
	}

	return nil
}

func (c *ClusterManager) InstallCustomComponents(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error {
	if err := c.clusterClient.installCustomComponents(ctx, clusterSpec, cluster); err != nil {
		return err
	}
	fullLifecycleAPI := features.IsActive(features.FullLifecycleAPI())
	if fullLifecycleAPI {
		err := c.clusterClient.SetEksaControllerEnvVar(ctx, features.FullLifecycleAPIEnvVar, strconv.FormatBool(fullLifecycleAPI), cluster.KubeconfigFile)
		if err != nil {
			return err
		}
	}
	return provider.InstallCustomProviderComponents(ctx, cluster.KubeconfigFile)
}

func (c *ClusterManager) CreateEKSANamespace(ctx context.Context, cluster *types.Cluster) error {
	return c.clusterClient.CreateNamespaceIfNotPresent(ctx, cluster.KubeconfigFile, constants.EksaSystemNamespace)
}

// CreateEKSAResources applies the eks-a cluster specs (cluster, datacenterconfig, machine configs, etc.), as well as the
// release bundle to the cluster. Before applying the spec, we pause eksa controller cluster and datacenter webhook validation
// so that the cluster spec can be created or updated in the cluster without webhook validation error.
func (c *ClusterManager) CreateEKSAResources(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec,
	datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig,
) error {
	if clusterSpec.Cluster.Namespace != "" {
		if err := c.clusterClient.CreateNamespaceIfNotPresent(ctx, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace); err != nil {
			return err
		}
	}

	clusterSpec.Cluster.PauseReconcile()
	datacenterConfig.PauseReconcile()

	resourcesSpec, err := clustermarshaller.MarshalClusterSpec(clusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		return err
	}
	logger.V(4).Info("Applying eksa yaml resources to cluster")
	logger.V(6).Info(string(resourcesSpec))
	if err = c.applyResource(ctx, cluster, resourcesSpec); err != nil {
		return err
	}
	return c.ApplyBundles(ctx, clusterSpec, cluster)
}

func (c *ClusterManager) ApplyBundles(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	bundleObj, err := yaml.Marshal(clusterSpec.Bundles)
	if err != nil {
		return fmt.Errorf("outputting bundle yaml: %v", err)
	}
	logger.V(1).Info("Applying Bundles to cluster")
	err = c.clusterClient.ApplyKubeSpecFromBytes(ctx, cluster, bundleObj)
	if err != nil {
		return fmt.Errorf("applying bundle spec: %v", err)
	}
	return nil
}

func (c *ClusterManager) PauseEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error {
	if clusterSpec.Cluster.IsSelfManaged() {
		return c.pauseEksaReconcileForManagementAndWorkloadClusters(ctx, cluster, clusterSpec, provider)
	}

	return c.pauseReconcileForCluster(ctx, cluster, clusterSpec.Cluster, provider)
}

func (c *ClusterManager) pauseEksaReconcileForManagementAndWorkloadClusters(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error {
	clusters := &v1alpha1.ClusterList{}
	err := c.clusterClient.ListObjects(ctx, eksaClusterResourceType, clusterSpec.Cluster.Namespace, managementCluster.KubeconfigFile, clusters)
	if err != nil {
		return err
	}

	for _, w := range clusters.Items {
		if w.ManagedBy() != clusterSpec.Cluster.Name {
			continue
		}

		if err := c.pauseReconcileForCluster(ctx, managementCluster, &w, provider); err != nil {
			return err
		}
	}

	return nil
}

func (c *ClusterManager) pauseReconcileForCluster(ctx context.Context, clusterCreds *types.Cluster, cluster *v1alpha1.Cluster, provider providers.Provider) error {
	pausedAnnotation := map[string]string{cluster.PausedAnnotation(): "true"}
	err := c.clusterClient.UpdateAnnotationInNamespace(ctx, provider.DatacenterResourceType(), cluster.Spec.DatacenterRef.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
	if err != nil {
		return fmt.Errorf("updating annotation when pausing datacenterconfig reconciliation: %v", err)
	}
	if provider.MachineResourceType() != "" {
		for _, machineConfigRef := range cluster.MachineConfigRefs() {
			err = c.clusterClient.UpdateAnnotationInNamespace(ctx, provider.MachineResourceType(), machineConfigRef.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
			if err != nil {
				return fmt.Errorf("updating annotation when pausing reconciliation for machine config %s: %v", machineConfigRef.Name, err)
			}
		}
	}

	err = c.clusterClient.UpdateAnnotationInNamespace(ctx, cluster.ResourceType(), cluster.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
	if err != nil {
		return fmt.Errorf("updating annotation when pausing cluster reconciliation: %v", err)
	}
	return nil
}

func (c *ClusterManager) resumeEksaReconcileForManagementAndWorkloadClusters(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error {
	clusters := &v1alpha1.ClusterList{}
	err := c.clusterClient.ListObjects(ctx, eksaClusterResourceType, clusterSpec.Cluster.Namespace, managementCluster.KubeconfigFile, clusters)
	if err != nil {
		return err
	}

	for _, w := range clusters.Items {
		if w.ManagedBy() != clusterSpec.Cluster.Name {
			continue
		}

		if err := c.resumeReconcileForCluster(ctx, managementCluster, &w, provider); err != nil {
			return err
		}
	}

	return nil
}

func (c *ClusterManager) ResumeEKSAControllerReconcile(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error {
	// clear pause annotation
	clusterSpec.Cluster.ClearPauseAnnotation()
	provider.DatacenterConfig(clusterSpec).ClearPauseAnnotation()

	if clusterSpec.Cluster.IsSelfManaged() {
		return c.resumeEksaReconcileForManagementAndWorkloadClusters(ctx, cluster, clusterSpec, provider)
	}

	return c.resumeReconcileForCluster(ctx, cluster, clusterSpec.Cluster, provider)
}

func (c *ClusterManager) resumeReconcileForCluster(ctx context.Context, clusterCreds *types.Cluster, cluster *v1alpha1.Cluster, provider providers.Provider) error {
	pausedAnnotation := cluster.PausedAnnotation()
	err := c.clusterClient.RemoveAnnotationInNamespace(ctx, provider.DatacenterResourceType(), cluster.Spec.DatacenterRef.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
	if err != nil {
		return fmt.Errorf("updating annotation when unpausing datacenterconfig reconciliation: %v", err)
	}
	if provider.MachineResourceType() != "" {
		for _, machineConfigRef := range cluster.MachineConfigRefs() {
			err = c.clusterClient.RemoveAnnotationInNamespace(ctx, provider.MachineResourceType(), machineConfigRef.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
			if err != nil {
				return fmt.Errorf("updating annotation when resuming reconciliation for machine config %s: %v", machineConfigRef.Name, err)
			}
		}
	}

	err = c.clusterClient.RemoveAnnotationInNamespace(ctx, cluster.ResourceType(), cluster.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
	if err != nil {
		return fmt.Errorf("updating annotation when unpausing cluster reconciliation: %v", err)
	}

	return nil
}

func (c *ClusterManager) applyResource(ctx context.Context, cluster *types.Cluster, resourcesSpec []byte) error {
	err := c.clusterClient.ApplyKubeSpecFromBytesForce(ctx, cluster, resourcesSpec)
	if err != nil {
		return fmt.Errorf("applying eks-a spec: %v", err)
	}
	return nil
}

func (c *ClusterManager) GetCurrentClusterSpec(ctx context.Context, clus *types.Cluster, clusterName string) (*cluster.Spec, error) {
	eksaCluster, err := c.clusterClient.GetEksaCluster(ctx, clus, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed getting EKS-A cluster to build current cluster Spec: %v", err)
	}

	return c.buildSpecForCluster(ctx, clus, eksaCluster)
}

func (c *ClusterManager) buildSpecForCluster(ctx context.Context, clus *types.Cluster, eksaCluster *v1alpha1.Cluster) (*cluster.Spec, error) {
	return cluster.BuildSpecForCluster(ctx, eksaCluster, c.bundlesFetcher(clus), c.eksdReleaseFetcher(clus), c.gitOpsFetcher(clus), c.fluxConfigFetcher(clus), c.oidcFetcher(clus), c.awsIamConfigFetcher(clus))
}

func (c *ClusterManager) bundlesFetcher(cluster *types.Cluster) cluster.BundlesFetch {
	return func(ctx context.Context, name, namespace string) (*releasev1alpha1.Bundles, error) {
		return c.clusterClient.GetBundles(ctx, cluster.KubeconfigFile, name, namespace)
	}
}

func (c *ClusterManager) eksdReleaseFetcher(cluster *types.Cluster) cluster.EksdReleaseFetch {
	return func(ctx context.Context, name, namespace string) (*eksdv1alpha1.Release, error) {
		return c.clusterClient.GetEksdRelease(ctx, name, namespace, cluster.KubeconfigFile)
	}
}

func (c *ClusterManager) gitOpsFetcher(cluster *types.Cluster) cluster.GitOpsFetch {
	return func(ctx context.Context, name, namespace string) (*v1alpha1.GitOpsConfig, error) {
		return c.clusterClient.GetEksaGitOpsConfig(ctx, name, cluster.KubeconfigFile, namespace)
	}
}

func (c *ClusterManager) fluxConfigFetcher(cluster *types.Cluster) cluster.FluxConfigFetch {
	return func(ctx context.Context, name, namespace string) (*v1alpha1.FluxConfig, error) {
		return c.clusterClient.GetEksaFluxConfig(ctx, name, cluster.KubeconfigFile, namespace)
	}
}

func (c *ClusterManager) oidcFetcher(cluster *types.Cluster) cluster.OIDCFetch {
	return func(ctx context.Context, name, namespace string) (*v1alpha1.OIDCConfig, error) {
		return c.clusterClient.GetEksaOIDCConfig(ctx, name, cluster.KubeconfigFile, namespace)
	}
}

func (c *ClusterManager) awsIamConfigFetcher(cluster *types.Cluster) cluster.AWSIamConfigFetch {
	return func(ctx context.Context, name, namespace string) (*v1alpha1.AWSIamConfig, error) {
		return c.clusterClient.GetEksaAWSIamConfig(ctx, name, cluster.KubeconfigFile, namespace)
	}
}

func (c *ClusterManager) DeleteGitOpsConfig(ctx context.Context, managementCluster *types.Cluster, name string, namespace string) error {
	return c.clusterClient.DeleteGitOpsConfig(ctx, managementCluster, name, namespace)
}

func (c *ClusterManager) DeleteOIDCConfig(ctx context.Context, managementCluster *types.Cluster, name string, namespace string) error {
	return c.clusterClient.DeleteOIDCConfig(ctx, managementCluster, name, namespace)
}

func (c *ClusterManager) DeleteAWSIamConfig(ctx context.Context, managementCluster *types.Cluster, name string, namespace string) error {
	return c.clusterClient.DeleteAWSIamConfig(ctx, managementCluster, name, namespace)
}

func (c *ClusterManager) DeleteEKSACluster(ctx context.Context, managementCluster *types.Cluster, name string, namespace string) error {
	return c.clusterClient.DeleteEKSACluster(ctx, managementCluster, name, namespace)
}

func (c *ClusterManager) DeletePackageResources(ctx context.Context, managementCluster *types.Cluster, clusterName string) error {
	return c.clusterClient.DeletePackageResources(ctx, managementCluster, clusterName)
}
