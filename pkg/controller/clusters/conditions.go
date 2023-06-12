package clusters

import (
	"context"
	"sync"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/errors"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	ciliumreconciler "github.com/aws/eks-anywhere/pkg/networking/cilium/reconciler"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultMaxConditionCheckJobs = 5
)

// ConditionChecker checks the Cluster state and returns a condition.
type ConditionChecker func(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) (*clusterv1.Condition, error)

// ConditionFetcher is composed of a set of ConditionChecker.
type ConditionFetcher []ConditionChecker

// Register registers checks with c.
func (cf *ConditionFetcher) Register(checkers ...ConditionChecker) {
	*cf = append(*cf, checkers...)
}

type conditionCheckerResult struct {
	condition *clusterv1.Condition
	err       error
}

func (cf *ConditionFetcher) run(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) <-chan conditionCheckerResult {
	results := make(chan conditionCheckerResult)
	checkers := make(chan ConditionChecker)
	wg := &sync.WaitGroup{}

	numWorkers := defaultMaxConditionCheckJobs
	if numWorkers > len(*cf) {
		numWorkers = len(*cf)
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for checker := range checkers {
				defer wg.Done()
				condition, err := checker(ctx, client, clusterSpec)
				results <- conditionCheckerResult{
					condition,
					err,
				}
			}
		}()
	}

	go func() {
		for _, c := range *cf {
			checkers <- c
		}
		close(checkers)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// RunAll runs all condition checkers concurrently and waits until they all finish,
// aggregating the errors if present.
func (cf *ConditionFetcher) RunAll(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) ([]*clusterv1.Condition, error) {
	var errList []error
	var conditions []*clusterv1.Condition

	for result := range cf.run(ctx, client, clusterSpec) {
		conditions = append(conditions, result.condition)
		errList = append(errList, result.err)
	}

	return conditions, errors.NewAggregate(errList)
}

// NewConditionFetcher creates a ConditionFetcher and any checkers passed will be registered.
func NewConditionFetcher(checkers ...ConditionChecker) *ConditionFetcher {
	var v ConditionFetcher
	v.Register(checkers...)
	return &v
}

// SetAllConditions sets all the given conditions on the provided Cluster.
func SetAllConditions(cluster *anywherev1.Cluster, conditionList []*clusterv1.Condition) {
	for _, c := range conditionList {
		conditions.Set(cluster, c)
	}
}

// CheckControlPlaneInitializedCondition updates the ControlPlaneInitialized condition if it hasn't already been set.
// This condition should be set only once.
func CheckControlPlaneInitializedCondition(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) (*clusterv1.Condition, error) {
	// Return early if the ControlPlaneInitializedCondition is already "True"
	if conditions.IsTrue(clusterSpec.Cluster, clusterv1.ControlPlaneInitializedCondition) {
		return nil, nil
	}
	// We can simply mirror this condition from the CAPI cluster
	condition, err := getConditionFromCAPICluster(ctx, client, clusterSpec.Cluster, clusterv1.ControlPlaneInitializedCondition)
	if err != nil {
		return nil, err
	}

	return condition, nil
}

// CheckControlPlaneReadyCondition updates the ControlPlaneReady condition based on the CAPI cluster condition
// and also the conditions on the control plane machines. The condition is marked "True", once all the
// requested control plane machines are ready.
func CheckControlPlaneReadyCondition(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) (*clusterv1.Condition, error) {
	capiClusterCondition, err := getConditionFromCAPICluster(ctx, client, clusterSpec.Cluster, clusterv1.ControlPlaneReadyCondition)
	if err != nil {
		return nil, err
	}

	// If the CAPICluster ControlPlaneReadyCondition is not "True", then we can assume that the control plane isn't ready and
	// Mirror the condition to the Cluster status. By checking the CAPI cluster first, we can capture all the exisitng reasons
	// that this may not be true first.
	if capiClusterCondition.Status != "True" {
		return capiClusterCondition, nil
	}

	// We want to ensure that the condition of the current cluster matches the input Cluster after checking with the CAPI cluster,
	// so, we do some further checks againts the Cluster machines to check if the expected control plane nodes have been actually rolled out
	// and ready.
	machines, err := controller.GetMachines(ctx, client, clusterSpec.Cluster)
	if err != nil {
		return nil, err
	}

	cpMachines := controller.ControlPlaneMachines(machines)

	expected := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count
	ready := countMachinesReady(cpMachines,
		controller.WithNodeRef(),
		controller.WithK8sVersion(clusterSpec.Cluster.Spec.KubernetesVersion),
		controller.WithConditionReady(),
		controller.WithNodeHealthy(),
	)

	if ready != expected {
		return conditions.FalseCondition(
			clusterv1.ControlPlaneReadyCondition,
			anywherev1.WaitingForControlPlaneNodesReadyReason,
			clusterv1.ConditionSeverityInfo, "Watiing for expected control plane nodes: %d replicas (ready %d)",
			expected,
			ready,
		), nil
	}

	return conditions.TrueCondition(clusterv1.ControlPlaneReadyCondition), nil
}

// CheckWorkersReadyCondition checks the WorkersReadyConditon condition based on the CAPI cluster condition
// and also the conditions on the worker machines. The condition is marked "True", once all the
// requested worker machines are ready.
func CheckWorkersReadyCondition(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) (*clusterv1.Condition, error) {
	machines, err := controller.GetMachines(ctx, client, clusterSpec.Cluster)
	if err != nil {
		return nil, err
	}

	workerMachines := controller.WorkerNodeMachines(machines)
	expected := 0
	for _, md := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		expected += *md.Count
	}

	ready := countMachinesReady(workerMachines,
		controller.WithNodeRef(),
		controller.WithK8sVersion(clusterSpec.Cluster.Spec.KubernetesVersion),
		controller.WithConditionReady(),
		controller.WithNodeHealthy(),
	)

	if ready != expected {
		return conditions.FalseCondition(
			anywherev1.WorkersReadyConditon,
			anywherev1.WaitingForWorkersReadyReason,
			clusterv1.ConditionSeverityInfo,
			"Waiting for expected workers nodes: %d replicas (ready %d)",
			expected,
			ready,
		), nil
	}

	return conditions.TrueCondition(anywherev1.WorkersReadyConditon), nil
}

// CheckDefaultCNIConfigured checks the DefaultCNIConfigured condition. The condition is marked "True" if
// the requested CNI is installed. client is connected to the target Kubernestes cluster, not the management cluster.
func CheckDefaultCNIConfigured(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) (*clusterv1.Condition, error) {
	installation, err := cilium.GetInstallation(ctx, client)
	if err != nil {
		return nil, err
	}

	clus := clusterSpec.Cluster
	ciliumCfg := clus.Spec.ClusterNetwork.CNIConfig.Cilium

	// If EKSA cilium is not installed and  EKS-A is responsible for the Cilium installation,
	// mark the DefaultCNIConfiguredCondition condition as "False".
	if !installation.Installed() && ciliumCfg.IsManaged() {
		return conditions.FalseCondition(
			anywherev1.DefaultCNIConfiguredCondition,
			anywherev1.WaitingForDefaultCNIConfiguredReason,
			clusterv1.ConditionSeverityInfo,
			"Waiting for default CNI to be configured",
		), nil
	}

	// If EKSA cilium is not installed and  EKS-A is not responsible for the Cilium installation,
	// Provided that cilium was installed before via checking a marker indicating this is an upgrade process
	// mark the DefaultCNIConfiguredCondition condition as "False" with a reason that default cni configuration was skipped.
	if !installation.Installed() && !ciliumCfg.IsManaged() && ciliumreconciler.CiliumWasInstalled(ctx, clus) {
		return conditions.FalseCondition(
			anywherev1.DefaultCNIConfiguredCondition,
			anywherev1.SkippedDefaultCNIConfigurationReason,
			clusterv1.ConditionSeverityInfo,
			"Skipped default CNI configuration",
		), nil
	}

	return conditions.TrueCondition(anywherev1.DefaultCNIConfiguredCondition), nil
}

// getConditionFromCAPICluster checks the condition on the CAPI cluster for the provided condition.
// If False, it mirrors the corresponding condition on the Cluster status.
// It returns a bool indicating that the condition was "True" or not
func getConditionFromCAPICluster(ctx context.Context, client client.Client, cluster *anywherev1.Cluster, clusterCondtion clusterv1.ConditionType) (*clusterv1.Condition, error) {
	noCAPIClusterCondition := conditions.FalseCondition(clusterCondtion, anywherev1.WaitingForCAPIClusterReason, clusterv1.ConditionSeverityInfo, "Waiting for CAPI cluster to be initialized")
	capiCluster, err := controller.GetCAPICluster(ctx, client, cluster)
	if err != nil {
		return noCAPIClusterCondition, err
	}

	if capiCluster == nil {
		return noCAPIClusterCondition, nil
	}

	condition := conditions.Get(capiCluster, clusterCondtion)
	if condition == nil {
		return conditions.FalseCondition(clusterCondtion, anywherev1.WaitingForCAPIClusterConditionReason, clusterv1.ConditionSeverityInfo, "Waiting for CAPI cluster to report %s condition", clusterCondtion), nil
	}

	return condition, nil
}

func countMachinesReady(machines []clusterv1.Machine, checkers ...controller.NodeReadyChecker) int {
	ready := 0
	for _, m := range machines {
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
	return ready
}
