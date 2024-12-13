package clustermanager

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/integer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
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
	kubectlResourceNotFoundRegex             = regexp.MustCompile(`.*the server doesn't have a resource type "(.*)".*`)
	eksaClusterResourceType                  = fmt.Sprintf("clusters.%s", v1alpha1.GroupVersion.Group)
)

type ClusterManager struct {
	eksaComponents     EKSAComponents
	ClientFactory      ClientFactory
	clusterClient      ClusterClient
	retrier            *retrier.Retrier
	writer             filewriter.FileWriter
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

type AwsIamAuth interface {
	CreateAndInstallAWSIAMAuthCASecret(ctx context.Context, managementCluster *types.Cluster, workloadClusterName string) error
	InstallAWSIAMAuth(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error
	UpgradeAWSIAMAuth(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec) error
	GenerateKubeconfig(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error
	GenerateManagementAWSIAMKubeconfig(ctx context.Context, cluster *types.Cluster) error
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
func New(client ClientFactory, clusterClient ClusterClient, writer filewriter.FileWriter, diagnosticBundleFactory diagnostics.DiagnosticBundleFactory, awsIamAuth AwsIamAuth, eksaComponents EKSAComponents, opts ...ClusterManagerOpt) *ClusterManager {
	c := &ClusterManager{
		eksaComponents:                   eksaComponents,
		ClientFactory:                    client,
		clusterClient:                    clusterClient,
		writer:                           writer,
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
		return true, exponentialRetryWaitTime(totalRetries)
	}
	return false, 0
}

func clusterctlMoveRetryPolicy(totalRetries int, err error) (retry bool, wait time.Duration) {
	// Retry only network errors.
	if match := clusterctlNetworkErrorRegex.MatchString(err.Error()); match {
		return true, exponentialRetryWaitTime(totalRetries)
	}
	return false, 0
}

func kubectlWaitRetryPolicy(totalRetries int, err error) (retry bool, wait time.Duration) {
	// Sometimes it is possible that the clusterctl move is successful,
	// but the clusters.cluster.x-k8s.io resource is not available on the cluster yet.
	//
	// Retry on transient 'server doesn't have a resource type' errors.
	// Use existing exponential backoff implementation for retry on these errors.
	if match := kubectlResourceNotFoundRegex.MatchString(err.Error()); match {
		return true, exponentialRetryWaitTime(totalRetries)
	}
	return false, 0
}

func exponentialRetryWaitTime(totalRetries int) time.Duration {
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
	r = retrier.New(c.clusterctlMoveTimeout, retrier.WithRetryPolicy(kubectlWaitRetryPolicy))
	err = r.Retry(func() error {
		return c.clusterClient.WaitForControlPlaneReady(ctx, to, c.controlPlaneWaitAfterMoveTimeout.String(), clusterName)
	})
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
		Data: map[string][]byte{
			"username": []byte(os.Getenv("REGISTRY_USERNAME")),
			"password": []byte(os.Getenv("REGISTRY_PASSWORD")),
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

// GenerateIamAuthKubeconfig generates a kubeconfig for interacting with the cluster with aws-iam-authenticator client.
func (c *ClusterManager) GenerateIamAuthKubeconfig(ctx context.Context, management, workload *types.Cluster, spec *cluster.Spec) error {
	return c.awsIamAuth.GenerateKubeconfig(ctx, management, workload, spec)
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

// Upgrade updates the eksa components in a cluster according to a Spec.
func (c *ClusterManager) Upgrade(ctx context.Context, cluster *types.Cluster, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	return c.eksaComponents.Upgrade(ctx, logger.Get(), cluster, currentManagementComponents, newManagementComponents, newSpec)
}

func (c *ClusterManager) CreateEKSANamespace(ctx context.Context, cluster *types.Cluster) error {
	return c.clusterClient.CreateNamespaceIfNotPresent(ctx, cluster.KubeconfigFile, constants.EksaSystemNamespace)
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

// ResumeEKSAControllerReconcile resumes a paused EKS-Anywhere cluster.
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

// AllowDeleteWhilePaused allows the deletion of paused clusters.
func (c *ClusterManager) AllowDeleteWhilePaused(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	return c.allowDeleteWhilePaused(ctx, cluster, clusterSpec.Cluster)
}

func (c *ClusterManager) allowDeleteWhilePaused(ctx context.Context, clusterCreds *types.Cluster, cluster *v1alpha1.Cluster) error {
	allowDelete := map[string]string{v1alpha1.AllowDeleteWhenPausedAnnotation: "true"}

	if err := c.clusterClient.UpdateAnnotationInNamespace(ctx, cluster.ResourceType(), cluster.Name, allowDelete, clusterCreds, cluster.Namespace); err != nil {
		return fmt.Errorf("updating paused annotation in cluster reconciliation: %v", err)
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

// GenerateAWSIAMKubeconfig generates a kubeconfig for interacting with the cluster with aws-iam-authenticator client.
func (c *ClusterManager) GenerateAWSIAMKubeconfig(ctx context.Context, cluster *types.Cluster) error {
	return c.awsIamAuth.GenerateManagementAWSIAMKubeconfig(ctx, cluster)
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
