package clustermanager

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/integer"
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
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	maxRetries             = 30
	defaultBackOffPeriod   = 5 * time.Second
	machineBackoff         = 1 * time.Second
	defaultMachinesMinWait = 30 * time.Minute

	// DefaultMaxWaitPerMachine is the default max time the cluster manager will wait per a machine.
	DefaultMaxWaitPerMachine = 10 * time.Minute
	// DefaultClusterWait is the default max time the cluster manager will wait for the capi cluster to be in ready state.
	DefaultClusterWait = 60 * time.Minute
	// DefaultControlPlaneWait is the default time the cluster manager will wait for the control plane to be ready.
	DefaultControlPlaneWait = 60 * time.Minute
	// DefaultControlPlaneWaitAfterMove is the default max time the cluster manager will wait for the control plane to be in ready state after the capi move operation.
	DefaultControlPlaneWaitAfterMove = 15 * time.Minute
	// DefaultDeploymentWait is the default max time the cluster manager will wait for the deployment to be available.
	DefaultDeploymentWait = 30 * time.Minute

	controlPlaneInProgressStr = "1m"
	etcdInProgressStr         = "1m"
	// DefaultEtcdWait is the default time the cluster manager will wait for ectd to be ready.
	DefaultEtcdWait = 60 * time.Minute
	// DefaultUnhealthyMachineTimeout is the default timeout for an unhealthy machine health check.
	DefaultUnhealthyMachineTimeout = 5 * time.Minute
	// DefaultNodeStartupTimeout is the default timeout for a machine without a node to be considered to have failed machine health check.
	DefaultNodeStartupTimeout = 10 * time.Minute
	// DefaultClusterctlMoveTimeout is arbitrarily established.  Equal to kubectl wait default timeouts.
	DefaultClusterctlMoveTimeout = 30 * time.Minute
)

var (
	clusterctlNetworkErrorRegex              = regexp.MustCompile(`.*failed to connect to the management cluster:.*`)
	clusterctlMoveProvisionedInfraErrorRegex = regexp.MustCompile(`.*failed to check for provisioned infrastructure*`)
	eksaClusterResourceType                  = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
)

type ClusterManager struct {
	eksaComponents     EKSAComponents
	ClientFactory      ClientFactory
	clusterClient      ClusterClient
	retrier            *retrier.Retrier
	writer             filewriter.FileWriter
	networking         Networking
	diagnosticsFactory diagnostics.DiagnosticBundleFactory
	awsIamAuth         AwsIamAuth

	machineMaxWait                   time.Duration
	machineBackoff                   time.Duration
	machinesMinWait                  time.Duration
	controlPlaneWaitTimeout          time.Duration
	controlPlaneWaitAfterMoveTimeout time.Duration
	externalEtcdWaitTimeout          time.Duration
	unhealthyMachineTimeout          time.Duration
	nodeStartupTimeout               time.Duration
	clusterWaitTimeout               time.Duration
	deploymentWaitTimeout            time.Duration
	clusterctlMoveTimeout            time.Duration
}

// ClientFactory builds Kubernetes clients.
type ClientFactory interface {
	// BuildClientFromKubeconfig builds a Kubernetes client from a kubeconfig file.
	BuildClientFromKubeconfig(kubeconfigPath string) (kubernetes.Client, error)
}

// CAPIClient performs operations on a cluster-api management cluster.
type CAPIClient interface {
	BackupManagement(ctx context.Context, cluster *types.Cluster, managementStatePath, clusterName string) error
	MoveManagement(ctx context.Context, from, target *types.Cluster, clusterName string) error
	InitInfrastructure(ctx context.Context, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error
	GetWorkloadKubeconfig(ctx context.Context, clusterName string, cluster *types.Cluster) ([]byte, error)
}

type Networking interface {
	Install(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, namespaces []string) error
	Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, namespaces []string) (*types.ChangeDiff, error)
	RunPostControlPlaneUpgradeSetup(ctx context.Context, cluster *types.Cluster) error
}

type AwsIamAuth interface {
	CreateAndInstallAWSIAMAuthCASecret(ctx context.Context, managementCluster *types.Cluster, workloadClusterName string) error
	InstallAWSIAMAuth(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error
	UpgradeAWSIAMAuth(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec) error
	GenerateKubeconfig(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error
}

// EKSAComponents allows to manage the eks-a components installation in a cluster.
type EKSAComponents interface {
	Install(ctx context.Context, log logr.Logger, cluster *types.Cluster, managementComponents *cluster.ManagementComponents, spec *cluster.Spec) error
	Upgrade(ctx context.Context, log logr.Logger, cluster *types.Cluster, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, newSpec *cluster.Spec) (*types.ChangeDiff, error)
}

type ClusterManagerOpt func(*ClusterManager)

// DefaultRetrier builds a retrier with the default configuration.
func DefaultRetrier() *retrier.Retrier {
	return retrier.NewWithMaxRetries(maxRetries, defaultBackOffPeriod)
}

// New constructs a new ClusterManager.
func New(client ClientFactory, clusterClient ClusterClient, networking Networking, writer filewriter.FileWriter, diagnosticBundleFactory diagnostics.DiagnosticBundleFactory, awsIamAuth AwsIamAuth, eksaComponents EKSAComponents, opts ...ClusterManagerOpt) *ClusterManager {
	c := &ClusterManager{
		eksaComponents:                   eksaComponents,
		ClientFactory:                    client,
		clusterClient:                    clusterClient,
		writer:                           writer,
		networking:                       networking,
		retrier:                          DefaultRetrier(),
		diagnosticsFactory:               diagnosticBundleFactory,
		machineMaxWait:                   DefaultMaxWaitPerMachine,
		machineBackoff:                   machineBackoff,
		machinesMinWait:                  defaultMachinesMinWait,
		awsIamAuth:                       awsIamAuth,
		controlPlaneWaitTimeout:          DefaultControlPlaneWait,
		controlPlaneWaitAfterMoveTimeout: DefaultControlPlaneWaitAfterMove,
		externalEtcdWaitTimeout:          DefaultEtcdWait,
		unhealthyMachineTimeout:          DefaultUnhealthyMachineTimeout,
		nodeStartupTimeout:               DefaultNodeStartupTimeout,
		clusterWaitTimeout:               DefaultClusterWait,
		deploymentWaitTimeout:            DefaultDeploymentWait,
		clusterctlMoveTimeout:            DefaultClusterctlMoveTimeout,
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

// WithUnhealthyMachineTimeout sets the timeout of an unhealthy machine health check.
func WithUnhealthyMachineTimeout(timeout time.Duration) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.unhealthyMachineTimeout = timeout
	}
}

// WithNodeStartupTimeout sets the timeout of a machine without a node to be considered to have failed machine health check.
func WithNodeStartupTimeout(timeout time.Duration) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.nodeStartupTimeout = timeout
	}
}

func WithRetrier(retrier *retrier.Retrier) ClusterManagerOpt {
	return func(c *ClusterManager) {
		c.retrier = retrier
	}
}

// WithNoTimeouts disables the timeout for all the waits and retries in cluster manager.
func WithNoTimeouts() ClusterManagerOpt {
	return func(c *ClusterManager) {
		noTimeoutRetrier := retrier.NewWithNoTimeout()
		maxTime := time.Duration(math.MaxInt64)

		c.retrier = noTimeoutRetrier
		c.machinesMinWait = maxTime
		c.controlPlaneWaitTimeout = maxTime
		c.controlPlaneWaitAfterMoveTimeout = maxTime
		c.externalEtcdWaitTimeout = maxTime
		c.clusterWaitTimeout = maxTime
		c.deploymentWaitTimeout = maxTime
		c.clusterctlMoveTimeout = maxTime
	}
}

func clusterctlMoveWaitForInfrastructureRetryPolicy(totalRetries int, err error) (retry bool, wait time.Duration) {
	// Retry both network and cluster move errors.
	if match := (clusterctlNetworkErrorRegex.MatchString(err.Error()) || clusterctlMoveProvisionedInfraErrorRegex.MatchString(err.Error())); match {
		return true, clusterctlMoveRetryWaitTime(totalRetries)
	}
	return false, 0
}

func clusterctlMoveRetryPolicy(totalRetries int, err error) (retry bool, wait time.Duration) {
	// Retry only network errors.
	if match := clusterctlNetworkErrorRegex.MatchString(err.Error()); match {
		return true, clusterctlMoveRetryWaitTime(totalRetries)
	}
	return false, 0
}

func clusterctlMoveRetryWaitTime(totalRetries int) time.Duration {
	// Exponential backoff on errors.  Retrier built-in backoff is linear, so implementing here.

	// Retrier first calls the policy before retry #1.  We want it zero-based for exponentiation.
	if totalRetries < 1 {
		totalRetries = 1
	}

	const networkFaultBaseRetryTime = 10 * time.Second
	const backoffFactor = 1.5

	return time.Duration(float64(networkFaultBaseRetryTime) * math.Pow(backoffFactor, float64(totalRetries-1)))
}

// BackupCAPI takes backup of management cluster's resources during the upgrade process.
func (c *ClusterManager) BackupCAPI(ctx context.Context, cluster *types.Cluster, managementStatePath, clusterName string) error {
	// Network errors, most commonly connection refused or timeout, can occur if either source
	// cluster becomes inaccessible during the move operation.  If this occurs without retries, clusterctl
	// abandons the move operation, and fails cluster upgrade.
	// Retrying once connectivity is re-established completes the partial move.
	// Here we use a retrier, with the above defined clusterctlMoveRetryPolicy policy, to attempt to
	// wait out the network disruption and complete the move.
	// Keeping clusterctlMoveTimeout to the same as MoveManagement since both uses the same command with the differrent params.

	r := retrier.New(c.clusterctlMoveTimeout, retrier.WithRetryPolicy(clusterctlMoveRetryPolicy))
	return c.backupCAPI(ctx, cluster, managementStatePath, clusterName, r)
}

// BackupCAPIWaitForInfrastructure takes backup of bootstrap cluster's resources during the upgrade process
// like BackupCAPI but with a retry policy to wait for infrastructure provisioning in addition to network errors.
func (c *ClusterManager) BackupCAPIWaitForInfrastructure(ctx context.Context, cluster *types.Cluster, managementStatePath, clusterName string) error {
	r := retrier.New(c.clusterctlMoveTimeout, retrier.WithRetryPolicy(clusterctlMoveWaitForInfrastructureRetryPolicy))
	return c.backupCAPI(ctx, cluster, managementStatePath, clusterName, r)
}

func (c *ClusterManager) backupCAPI(ctx context.Context, cluster *types.Cluster, managementStatePath, clusterName string, retrier *retrier.Retrier) error {
	err := retrier.Retry(func() error {
		return c.clusterClient.BackupManagement(ctx, cluster, managementStatePath, clusterName)
	})
	if err != nil {
		return fmt.Errorf("backing up CAPI resources of management cluster before moving to bootstrap cluster: %v", err)
	}
	return nil
}

func (c *ClusterManager) MoveCAPI(ctx context.Context, from, to *types.Cluster, clusterName string, clusterSpec *cluster.Spec, checkers ...types.NodeReadyChecker) error {
	logger.V(3).Info("Waiting for management machines to be ready before move")
	labels := []string{clusterv1.MachineControlPlaneNameLabel, clusterv1.MachineDeploymentNameLabel}
	if err := c.waitForNodesReady(ctx, from, clusterName, labels, checkers...); err != nil {
		return err
	}

	logger.V(3).Info("Waiting for management cluster to be ready before move")
	if err := c.clusterClient.WaitForClusterReady(ctx, from, c.clusterWaitTimeout.String(), clusterName); err != nil {
		return err
	}

	// Network errors, most commonly connection refused or timeout, can occur if either source or target
	// cluster becomes inaccessible during the move operation.  If this occurs without retries, clusterctl
	// abandons the move operation, leaving an unpredictable subset of the CAPI components copied to target
	// or deleted from source.  Retrying once connectivity is re-established completes the partial move.
	// Here we use a retrier, with the above defined clusterctlMoveRetryPolicy policy, to attempt to
	// wait out the network disruption and complete the move.

	r := retrier.New(c.clusterctlMoveTimeout, retrier.WithRetryPolicy(clusterctlMoveRetryPolicy))
	err := r.Retry(func() error {
		return c.clusterClient.MoveManagement(ctx, from, to, clusterName)
	})
	if err != nil {
		return fmt.Errorf("moving CAPI management from source to target: %v", err)
	}

	logger.V(3).Info("Waiting for workload cluster control plane to be ready after move")
	err = c.clusterClient.WaitForControlPlaneReady(ctx, to, c.controlPlaneWaitAfterMoveTimeout.String(), clusterName)
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
	clusterName := clusterSpec.Cluster.Name

	workloadCluster := &types.Cluster{
		Name: clusterName,
	}

	if err := c.applyProviderManifests(ctx, clusterSpec, managementCluster, provider); err != nil {
		return nil, err
	}

	if err := c.waitUntilControlPlaneAvailable(ctx, clusterSpec, managementCluster); err != nil {
		return nil, err
	}

	logger.V(3).Info("Waiting for workload kubeconfig generation", "cluster", clusterName)

	// Use a buffer to cache the kubeconfig.
	var buf bytes.Buffer

	if err := c.getWorkloadClusterKubeconfig(ctx, clusterName, managementCluster, &buf); err != nil {
		return nil, fmt.Errorf("waiting for workload kubeconfig: %v", err)
	}

	rawKubeconfig := buf.Bytes()

	// The Docker provider wants to update the kubeconfig to patch the server address before
	// we write it to disk. This is to ensure we can communicate with the cluster even when
	// hosted inside a Docker Desktop VM.
	if err := provider.UpdateKubeConfig(&rawKubeconfig, clusterName); err != nil {
		return nil, err
	}

	kubeconfigFile, err := c.writer.Write(
		kubeconfig.FormatWorkloadClusterKubeconfigFilename(clusterName),
		rawKubeconfig,
		filewriter.PersistentFile,
		filewriter.Permission0600,
	)
	if err != nil {
		return nil, fmt.Errorf("writing workload kubeconfig: %v", err)
	}
	workloadCluster.KubeconfigFile = kubeconfigFile

	return workloadCluster, nil
}

func (c *ClusterManager) waitUntilControlPlaneAvailable(
	ctx context.Context,
	clusterSpec *cluster.Spec,
	managementCluster *types.Cluster,
) error {
	// If we have external etcd we need to wait for that first as control plane nodes can't
	// come up without it.
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		logger.V(3).Info("Waiting for external etcd to be ready", "cluster", clusterSpec.Cluster.Name)
		err := c.clusterClient.WaitForManagedExternalEtcdReady(
			ctx,
			managementCluster,
			c.externalEtcdWaitTimeout.String(),
			clusterSpec.Cluster.Name,
		)
		if err != nil {
			return fmt.Errorf("waiting for external etcd for workload cluster to be ready: %v", err)
		}
		logger.V(3).Info("External etcd is ready")
	}

	logger.V(3).Info("Waiting for control plane to be available")
	err := c.clusterClient.WaitForControlPlaneAvailable(
		ctx,
		managementCluster,
		c.controlPlaneWaitTimeout.String(),
		clusterSpec.Cluster.Name,
	)
	if err != nil {
		return fmt.Errorf("waiting for control plane to be ready: %v", err)
	}

	return nil
}

func (c *ClusterManager) applyProviderManifests(
	ctx context.Context,
	spec *cluster.Spec,
	management *types.Cluster,
	provider providers.Provider,
) error {
	cpContent, mdContent, err := provider.GenerateCAPISpecForCreate(ctx, management, spec)
	if err != nil {
		return fmt.Errorf("generating capi spec: %v", err)
	}

	content := templater.AppendYamlResources(cpContent, mdContent)

	if err = c.writeCAPISpecFile(spec.Cluster.Name, content); err != nil {
		return err
	}

	err = c.clusterClient.ApplyKubeSpecFromBytesWithNamespace(ctx, management, content, constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("applying capi spec: %v", err)
	}

	return nil
}

func (c *ClusterManager) getWorkloadClusterKubeconfig(ctx context.Context, clusterName string, managementCluster *types.Cluster, w io.Writer) error {
	kubeconfig, err := c.clusterClient.GetWorkloadKubeconfig(ctx, clusterName, managementCluster)
	if err != nil {
		return fmt.Errorf("getting workload kubeconfig: %v", err)
	}

	if _, err := io.Copy(w, bytes.NewReader(kubeconfig)); err != nil {
		return err
	}

	return nil
}

func (c *ClusterManager) RunPostCreateWorkloadCluster(ctx context.Context, managementCluster, workloadCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.V(3).Info("Waiting for controlplane and worker machines to be ready")
	labels := []string{clusterv1.MachineControlPlaneNameLabel, clusterv1.MachineDeploymentNameLabel}
	return c.waitForNodesReady(ctx, managementCluster, workloadCluster.Name, labels, types.WithNodeRef(), types.WithNodeHealthy())
}

func (c *ClusterManager) DeleteCluster(ctx context.Context, managementCluster, clusterToDelete *types.Cluster, provider providers.Provider, clusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.IsManaged() {
		if err := c.deleteEKSAObjects(ctx, managementCluster, clusterToDelete, provider, clusterSpec); err != nil {
			return err
		}
	}

	logger.V(1).Info("Deleting CAPI cluster", "name", clusterToDelete.Name)
	if err := c.clusterClient.DeleteCluster(ctx, managementCluster, clusterToDelete); err != nil {
		return err
	}

	return provider.PostClusterDeleteValidate(ctx, managementCluster)
}

func (c *ClusterManager) deleteEKSAObjects(ctx context.Context, managementCluster, clusterToDelete *types.Cluster, provider providers.Provider, clusterSpec *cluster.Spec) error {
	log := logger.Get()
	log.V(1).Info("Deleting EKS-A objects", "cluster", clusterSpec.Cluster.Name)

	log.V(2).Info("Pausing EKS-A reconciliation", "cluster", clusterSpec.Cluster.Name)
	if err := c.PauseEKSAControllerReconcile(ctx, clusterToDelete, clusterSpec, provider); err != nil {
		return err
	}

	log.V(2).Info("Deleting EKS-A Cluster", "name", clusterSpec.Cluster.Name)
	if err := c.clusterClient.DeleteEKSACluster(ctx, managementCluster, clusterSpec.Cluster.Name, clusterSpec.Cluster.Namespace); err != nil {
		return err
	}

	if clusterSpec.GitOpsConfig != nil {
		log.V(2).Info("Deleting GitOpsConfig", "name", clusterSpec.GitOpsConfig.Name)
		if err := c.clusterClient.DeleteGitOpsConfig(ctx, managementCluster, clusterSpec.GitOpsConfig.Name, clusterSpec.GitOpsConfig.Namespace); err != nil {
			return err
		}
	}

	if clusterSpec.OIDCConfig != nil {
		log.V(2).Info("Deleting OIDCConfig", "name", clusterSpec.OIDCConfig.Name)
		if err := c.clusterClient.DeleteOIDCConfig(ctx, managementCluster, clusterSpec.OIDCConfig.Name, clusterSpec.OIDCConfig.Namespace); err != nil {
			return err
		}
	}

	if clusterSpec.AWSIamConfig != nil {
		log.V(2).Info("Deleting AWSIamConfig", "name", clusterSpec.AWSIamConfig.Name)
		if err := c.clusterClient.DeleteAWSIamConfig(ctx, managementCluster, clusterSpec.AWSIamConfig.Name, clusterSpec.AWSIamConfig.Namespace); err != nil {
			return err
		}
	}

	log.V(2).Info("Cleaning up provider specific resources")
	if err := provider.DeleteResources(ctx, clusterSpec); err != nil {
		return err
	}

	return nil
}

func (c *ClusterManager) UpgradeCluster(ctx context.Context, managementCluster, workloadCluster *types.Cluster, newClusterSpec *cluster.Spec, provider providers.Provider) error {
	eksaMgmtCluster := workloadCluster
	if managementCluster != nil {
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
	if err = c.waitForNodesReady(ctx, managementCluster, newClusterSpec.Cluster.Name, []string{clusterv1.MachineControlPlaneNameLabel}, types.WithNodeRef(), types.WithNodeHealthy()); err != nil {
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
	if err = c.waitForNodesReady(ctx, managementCluster, newClusterSpec.Cluster.Name, []string{clusterv1.MachineDeploymentNameLabel}, types.WithNodeRef(), types.WithNodeHealthy()); err != nil {
		return err
	}

	logger.V(3).Info("Waiting for workload cluster capi components to be ready after upgrade")
	err = c.waitForCAPI(ctx, eksaMgmtCluster, provider, externalEtcdTopology)
	if err != nil {
		return fmt.Errorf("waiting for workload cluster capi components to be ready: %v", err)
	}

	if newClusterSpec.AWSIamConfig != nil {
		logger.V(3).Info("Run aws-iam-authenticator upgrade operations")
		if err = c.awsIamAuth.UpgradeAWSIAMAuth(ctx, workloadCluster, newClusterSpec); err != nil {
			return fmt.Errorf("running aws-iam-authenticator upgrade operations: %v", err)
		}
	}

	return nil
}

// EKSAClusterSpecChanged checks if a cluster's new Spec is different from the current Spec.
func (c *ClusterManager) EKSAClusterSpecChanged(ctx context.Context, clus *types.Cluster, newClusterSpec *cluster.Spec) (bool, error) {
	cc, err := c.clusterClient.GetEksaCluster(ctx, clus, newClusterSpec.Cluster.Name)
	if err != nil {
		return false, err
	}

	if !cc.Equal(newClusterSpec.Cluster) {
		logger.V(3).Info("Existing cluster and new cluster spec differ")
		return true, nil
	}

	currentClusterSpec, err := c.buildSpecForCluster(ctx, clus, cc)
	if err != nil {
		return false, err
	}

	changed, err := compareEKSAClusterSpec(ctx, currentClusterSpec, newClusterSpec)
	if err != nil {
		return false, err
	}

	if !changed {
		logger.V(3).Info("Clusters are the same")
	}
	return changed, nil
}

func compareEKSAClusterSpec(ctx context.Context, currentClusterSpec, newClusterSpec *cluster.Spec) (bool, error) {
	newVersionsBundle := newClusterSpec.RootVersionsBundle()
	oldVersionsBundle := currentClusterSpec.RootVersionsBundle()

	if oldVersionsBundle.EksD.Name != newVersionsBundle.EksD.Name {
		logger.V(3).Info("New eks-d release detected")
		return true, nil
	}

	for _, workerNode := range newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		newVersionsBundle = newClusterSpec.WorkerNodeGroupVersionsBundle(workerNode)
		oldVersionsBundle = currentClusterSpec.WorkerNodeGroupVersionsBundle(workerNode)

		if oldVersionsBundle.EksD.Name != newVersionsBundle.EksD.Name {
			logger.V(3).Info("New eks-d release detected")
			return true, nil
		}
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

	return false, nil
}

// CreateRegistryCredSecret creates the registry-credentials secret on a managment cluster.
func (c *ClusterManager) CreateRegistryCredSecret(ctx context.Context, mgmt *types.Cluster) error {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
			Name:      "registry-credentials",
		},
		StringData: map[string]string{
			"username": os.Getenv("REGISTRY_USERNAME"),
			"password": os.Getenv("REGISTRY_PASSWORD"),
		},
	}

	return c.clusterClient.Apply(ctx, mgmt.KubeconfigFile, secret)
}

// InstallCAPI installs the cluster-api components in a cluster.
func (c *ClusterManager) InstallCAPI(ctx context.Context, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error {
	err := c.clusterClient.InitInfrastructure(ctx, managementComponents, clusterSpec, cluster, provider)
	if err != nil {
		return fmt.Errorf("initializing capi resources in cluster: %v", err)
	}

	return c.waitForCAPI(ctx, cluster, provider, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil)
}

func (c *ClusterManager) waitForCAPI(ctx context.Context, cluster *types.Cluster, provider providers.Provider, externalEtcdTopology bool) error {
	err := c.waitForDeployments(ctx, internal.CAPIDeployments, cluster, c.deploymentWaitTimeout.String())
	if err != nil {
		return err
	}

	if externalEtcdTopology {
		err := c.waitForDeployments(ctx, internal.ExternalEtcdDeployments, cluster, c.deploymentWaitTimeout.String())
		if err != nil {
			return err
		}
	}

	err = c.waitForDeployments(ctx, provider.GetDeployments(), cluster, c.deploymentWaitTimeout.String())
	if err != nil {
		return err
	}

	return nil
}

func (c *ClusterManager) waitForDeployments(ctx context.Context, deploymentsByNamespace map[string][]string, cluster *types.Cluster, timeout string) error {
	for namespace, deployments := range deploymentsByNamespace {
		for _, deployment := range deployments {
			err := c.clusterClient.WaitForDeployment(ctx, cluster, timeout, "Available", deployment, namespace)
			if err != nil {
				return fmt.Errorf("waiting for %s in namespace %s: %v", deployment, namespace, err)
			}
		}
	}
	return nil
}

func (c *ClusterManager) InstallNetworking(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error {
	return c.networking.Install(ctx, cluster, clusterSpec, getProviderNamespaces(provider.GetDeployments()))
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

// InstallMachineHealthChecks installs machine health checks for a given eksa cluster.
func (c *ClusterManager) InstallMachineHealthChecks(ctx context.Context, clusterSpec *cluster.Spec, workloadCluster *types.Cluster) error {
	mhc, err := templater.ObjectsToYaml(kubernetes.ObjectsToRuntimeObjects(clusterapi.MachineHealthCheckObjects(clusterSpec.Cluster))...)
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
func (c *ClusterManager) InstallAwsIamAuth(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error {
	return c.awsIamAuth.InstallAWSIAMAuth(ctx, management, workload, spec)
}

// GenerateIamAuthKubeconfig generates a kubeconfig for interacting with the cluster with aws-iam-authenticator client.
func (c *ClusterManager) GenerateIamAuthKubeconfig(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error {
	return c.awsIamAuth.GenerateKubeconfig(ctx, management, workload, spec)
}

func (c *ClusterManager) CreateAwsIamAuthCaSecret(ctx context.Context, managementCluster *types.Cluster, workloadClusterName string) error {
	return c.awsIamAuth.CreateAndInstallAWSIAMAuthCASecret(ctx, managementCluster, workloadClusterName)
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

	timeout := c.totalTimeoutForMachinesReadyWait(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count)
	r := retrier.New(timeout)
	if err := r.Retry(isCpReady); err != nil {
		return fmt.Errorf("retries exhausted waiting for controlplane replicas to be ready: %v", err)
	}
	return nil
}

func (c *ClusterManager) waitForMachineDeploymentReplicasReady(ctx context.Context, managementCluster *types.Cluster, clusterSpec *cluster.Spec) error {
	ready, total := 0, 0
	policy := func(_ int, _ error) (bool, time.Duration) {
		return true, c.machineBackoff * time.Duration(integer.IntMax(1, total-ready))
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

	timeout := c.totalTimeoutForMachinesReadyWait(machineDeploymentReplicasCount)
	r := retrier.New(timeout, retrier.WithRetryPolicy(policy))
	if err := r.Retry(areMdReplicasReady); err != nil {
		return fmt.Errorf("retries exhausted waiting for machinedeployment replicas to be ready: %v", err)
	}
	return nil
}

// totalTimeoutForMachinesReadyWait calculates the total timeout when waiting for machines to be ready.
// The timeout increases linearly with the number of machines but can never be less than the configured
// minimun.
func (c *ClusterManager) totalTimeoutForMachinesReadyWait(replicaCount int) time.Duration {
	timeout := time.Duration(replicaCount) * c.machineMaxWait
	if timeout <= c.machinesMinWait {
		timeout = c.machinesMinWait
	}

	return timeout
}

func (c *ClusterManager) waitForNodesReady(ctx context.Context, managementCluster *types.Cluster, clusterName string, labels []string, checkers ...types.NodeReadyChecker) error {
	totalNodes, err := c.getNodesCount(ctx, managementCluster, clusterName, labels)
	if err != nil {
		return fmt.Errorf("getting the total count of nodes: %v", err)
	}

	readyNodes := 0
	policy := func(_ int, _ error) (bool, time.Duration) {
		return true, c.machineBackoff * time.Duration(integer.IntMax(1, totalNodes-readyNodes))
	}

	areNodesReady := func() error {
		var err error
		readyNodes, err = c.countNodesReady(ctx, managementCluster, clusterName, labels, checkers...)
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

	err = areNodesReady()
	if err == nil {
		return nil
	}

	timeout := c.totalTimeoutForMachinesReadyWait(totalNodes)
	r := retrier.New(timeout, retrier.WithRetryPolicy(policy))
	if err := r.Retry(areNodesReady); err != nil {
		return fmt.Errorf("retries exhausted waiting for machines to be ready: %v", err)
	}

	return nil
}

func (c *ClusterManager) getNodesCount(ctx context.Context, managementCluster *types.Cluster, clusterName string, labels []string) (int, error) {
	totalNodes := 0

	labelsMap := make(map[string]interface{}, len(labels))
	for _, label := range labels {
		labelsMap[label] = nil
	}

	if _, ok := labelsMap[clusterv1.MachineControlPlaneNameLabel]; ok {
		kcp, err := c.clusterClient.GetKubeadmControlPlane(ctx, managementCluster, clusterName, executables.WithCluster(managementCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return 0, fmt.Errorf("getting KubeadmControlPlane for cluster %s: %v", clusterName, err)
		}
		totalNodes += int(*kcp.Spec.Replicas)
	}

	if _, ok := labelsMap[clusterv1.MachineDeploymentNameLabel]; ok {
		mds, err := c.clusterClient.GetMachineDeploymentsForCluster(ctx, clusterName, executables.WithCluster(managementCluster), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			return 0, fmt.Errorf("getting KubeadmControlPlane for cluster %s: %v", clusterName, err)
		}
		for _, md := range mds {
			totalNodes += int(*md.Spec.Replicas)
		}
	}

	return totalNodes, nil
}

func (c *ClusterManager) countNodesReady(ctx context.Context, managementCluster *types.Cluster, clusterName string, labels []string, checkers ...types.NodeReadyChecker) (ready int, err error) {
	machines, err := c.clusterClient.GetMachines(ctx, managementCluster, clusterName)
	if err != nil {
		return 0, fmt.Errorf("getting machines resources from management cluster: %v", err)
	}

	for _, m := range machines {
		// Extracted from cluster-api: NodeRef is considered a better signal than InfrastructureReady,
		// because it ensures the node in the workload cluster is up and running.
		if !m.HasAnyLabel(labels) {
			continue
		}

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
	return ready, nil
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
		mdName := clusterapi.MachineDeploymentName(newSpec.Cluster, group)
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

// InstallCustomComponents installs the eks-a components in a cluster.
func (c *ClusterManager) InstallCustomComponents(ctx context.Context, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error {
	if err := c.eksaComponents.Install(ctx, logger.Get(), cluster, managementComponents, clusterSpec); err != nil {
		return err
	}

	// TODO(g-gaston): should this be moved inside the components installer?
	return provider.InstallCustomProviderComponents(ctx, cluster.KubeconfigFile)
}

// Upgrade updates the eksa components in a cluster according to a Spec.
func (c *ClusterManager) Upgrade(ctx context.Context, cluster *types.Cluster, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	return c.eksaComponents.Upgrade(ctx, logger.Get(), cluster, currentManagementComponents, newManagementComponents, newSpec)
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
	if err = c.ApplyBundles(ctx, clusterSpec, cluster); err != nil {
		return err
	}
	return c.ApplyReleases(ctx, clusterSpec, cluster)
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

	// We need to update this config map with the new upgrader images whenever we
	// apply a new Bundles object to the cluster in order to support in-place upgrades.
	cm, err := c.getUpgraderImagesFromBundle(ctx, cluster, clusterSpec)
	if err != nil {
		return fmt.Errorf("getting upgrader images from bundle: %v", err)
	}
	if err = c.clusterClient.Apply(ctx, cluster.KubeconfigFile, cm); err != nil {
		return fmt.Errorf("applying upgrader images config map: %v", err)
	}
	return nil
}

// ApplyReleases applies the EKSARelease manifest.
func (c *ClusterManager) ApplyReleases(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	releaseObj, err := yaml.Marshal(clusterSpec.EKSARelease)
	if err != nil {
		return fmt.Errorf("outputting release yaml: %v", err)
	}
	logger.V(1).Info("Applying EKSARelease to cluster")
	err = c.clusterClient.ApplyKubeSpecFromBytes(ctx, cluster, releaseObj)
	if err != nil {
		return fmt.Errorf("applying release spec: %v", err)
	}
	return nil
}

// PauseCAPIWorkloadClusters pauses all workload CAPI clusters except the management cluster.
func (c *ClusterManager) PauseCAPIWorkloadClusters(ctx context.Context, managementCluster *types.Cluster) error {
	clusters, err := c.clusterClient.GetClusters(ctx, managementCluster)
	if err != nil {
		return err
	}

	for _, w := range clusters {
		// skip pausing management cluster
		if w.Metadata.Name == managementCluster.Name {
			continue
		}
		if err = c.clusterClient.PauseCAPICluster(ctx, w.Metadata.Name, managementCluster.KubeconfigFile); err != nil {
			return err
		}
	}
	return nil
}

// ResumeCAPIWorkloadClusters resumes all workload CAPI clusters except the management cluster.
func (c *ClusterManager) ResumeCAPIWorkloadClusters(ctx context.Context, managementCluster *types.Cluster) error {
	clusters, err := c.clusterClient.GetClusters(ctx, managementCluster)
	if err != nil {
		return err
	}

	for _, w := range clusters {
		// skip resuming management cluster
		if w.Metadata.Name == managementCluster.Name {
			continue
		}
		if err = c.clusterClient.ResumeCAPICluster(ctx, w.Metadata.Name, managementCluster.KubeconfigFile); err != nil {
			return err
		}
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
		return fmt.Errorf("updating paused annotation in cluster reconciliation: %v", err)
	}

	if err = c.clusterClient.UpdateAnnotationInNamespace(ctx,
		cluster.ResourceType(),
		cluster.Name,
		map[string]string{v1alpha1.ManagedByCLIAnnotation: "true"},
		clusterCreds,
		cluster.Namespace,
	); err != nil {
		return fmt.Errorf("updating managed by cli annotation in cluster when pausing cluster reconciliation: %v", err)
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
		return fmt.Errorf("removing paused annotation when resuming datacenterconfig reconciliation: %v", err)
	}

	if provider.MachineResourceType() != "" {
		for _, machineConfigRef := range cluster.MachineConfigRefs() {
			err = c.clusterClient.RemoveAnnotationInNamespace(ctx, provider.MachineResourceType(), machineConfigRef.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
			if err != nil {
				return fmt.Errorf("removing paused annotation when resuming reconciliation for machine config %s: %v", machineConfigRef.Name, err)
			}
		}
	}

	err = c.clusterClient.RemoveAnnotationInNamespace(ctx, cluster.ResourceType(), cluster.Name, pausedAnnotation, clusterCreds, cluster.Namespace)
	if err != nil {
		return fmt.Errorf("removing paused annotation when resuming cluster reconciliation: %v", err)
	}

	if err = c.clusterClient.RemoveAnnotationInNamespace(ctx,
		cluster.ResourceType(),
		cluster.Name,
		v1alpha1.ManagedByCLIAnnotation,
		clusterCreds,
		cluster.Namespace,
	); err != nil {
		return fmt.Errorf("removing managed by CLI annotation when resuming cluster reconciliation: %v", err)
	}

	return nil
}

// RemoveManagedByCLIAnnotationForCluster removes the managed-by-cli annotation from the cluster.
func (c *ClusterManager) RemoveManagedByCLIAnnotationForCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, provider providers.Provider) error {
	err := c.clusterClient.RemoveAnnotationInNamespace(ctx, clusterSpec.Cluster.ResourceType(), cluster.Name, v1alpha1.ManagedByCLIAnnotation, cluster, clusterSpec.Cluster.Namespace)
	if err != nil {
		return fmt.Errorf("removing managed by CLI annotation after apply cluster spec: %v", err)
	}
	return nil
}

func (c *ClusterManager) applyResource(ctx context.Context, cluster *types.Cluster, resourcesSpec []byte) error {
	err := c.clusterClient.ApplyKubeSpecFromBytes(ctx, cluster, resourcesSpec)
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
	client, err := c.ClientFactory.BuildClientFromKubeconfig(clus.KubeconfigFile)
	if err != nil {
		return nil, err
	}
	return cluster.BuildSpec(ctx, client, eksaCluster)
}

func (c *ClusterManager) DeletePackageResources(ctx context.Context, managementCluster *types.Cluster, clusterName string) error {
	return c.clusterClient.DeletePackageResources(ctx, managementCluster, clusterName)
}

// CreateNamespace creates a namespace on the target cluster if it does not already exist.
func (c *ClusterManager) CreateNamespace(ctx context.Context, targetCluster *types.Cluster, namespace string) error {
	return c.clusterClient.CreateNamespaceIfNotPresent(ctx, targetCluster.KubeconfigFile, namespace)
}

func (c *ClusterManager) getUpgraderImagesFromBundle(ctx context.Context, cluster *types.Cluster, cl *cluster.Spec) (*corev1.ConfigMap, error) {
	upgraderImages := make(map[string]string)
	for _, versionBundle := range cl.Bundles.Spec.VersionsBundles {
		eksD := versionBundle.EksD
		eksdVersion := fmt.Sprintf("%s-eks-%s-%s", eksD.KubeVersion, eksD.ReleaseChannel, strings.Split(eksD.Name, "-")[4])
		if _, ok := upgraderImages[eksdVersion]; !ok {
			upgraderImages[eksdVersion] = versionBundle.Upgrader.Upgrader.URI
		}
	}

	upgraderConfigMap, err := c.clusterClient.GetConfigMap(ctx, cluster.KubeconfigFile, constants.UpgraderConfigMapName, constants.EksaSystemNamespace)
	if err != nil {
		if executables.IsKubectlNotFoundError(err) {
			return newUpgraderConfigMap(upgraderImages), nil
		}
		return nil, err
	}

	for version, image := range upgraderImages {
		upgraderConfigMap.Data[version] = image
	}

	return upgraderConfigMap, nil
}

func newUpgraderConfigMap(m map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.UpgraderConfigMapName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: m,
	}
}
