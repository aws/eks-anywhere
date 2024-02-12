package clusters

import (
	"context"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

// UpdateClusterStatusForControlPlane checks the current state of the Cluster's control plane and updates the
// Cluster status information.
// There is a posibility that UpdateClusterStatusForControlPlane does not update the
// controleplane status specially in case where it still waiting for cluster objects to be created.
func UpdateClusterStatusForControlPlane(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) error {
	kcp, err := controller.GetKubeadmControlPlane(ctx, client, cluster)
	if err != nil {
		return errors.Wrapf(err, "getting kubeadmcontrolplane")
	}

	var etcdadmCluster *etcdv1.EtcdadmCluster
	if cluster.Spec.ExternalEtcdConfiguration != nil {
		capiCluster, err := controller.GetCAPICluster(ctx, client, cluster)
		if err != nil {
			return errors.Wrap(err, "getting capi cluster")
		}
		if capiCluster != nil {
			etcdadmCluster, err = getEtcdadmCluster(ctx, client, capiCluster)
			if err != nil {
				return errors.Wrap(err, "reading etcdadm cluster")
			}
		}
	}

	updateControlPlaneInitializedCondition(cluster, kcp)
	updateConditionsForEtcdAndControlPlane(cluster, kcp, etcdadmCluster)

	return nil
}

// UpdateClusterStatusForWorkers checks the current state of the Cluster's workers and updates the
// Cluster status information.
func UpdateClusterStatusForWorkers(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) error {
	machineDeployments, err := controller.GetMachineDeployments(ctx, client, cluster)
	if err != nil {
		return errors.Wrap(err, "getting machine deployments")
	}

	updateWorkersReadyCondition(cluster, machineDeployments)
	return nil
}

// UpdateClusterStatusForCNI updates the Cluster status for the default cni before the control plane is ready. The CNI reconciler
// handles the rest of the logic for determining the condition and updating the status based on the current state of the cluster.
func UpdateClusterStatusForCNI(ctx context.Context, cluster *anywherev1.Cluster) {
	// Here, we want to initialize the DefaultCNIConfigured condition only when the condition does not exist,
	// such as in the event of cluster creation. In this case, when the control plane is not ready, we can assume
	// the CNI is not ready yet.
	if !conditions.IsTrue(cluster, anywherev1.ControlPlaneReadyCondition) &&
		conditions.Get(cluster, anywherev1.DefaultCNIConfiguredCondition) == nil {
		conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.ControlPlaneNotReadyReason, clusterv1.ConditionSeverityInfo, "")
		return
	}
}

// updateConditionsForEtcdAndControlPlane updates the ControlPlaneReady condition if etcdadm cluster is not ready.
func updateConditionsForEtcdAndControlPlane(cluster *anywherev1.Cluster, kcp *controlplanev1.KubeadmControlPlane, etcdadmCluster *etcdv1.EtcdadmCluster) {
	// Make sure etcd cluster is ready before marking ControlPlaneReady status to true
	// This condition happens while creating a workload cluster from the management cluster using controller
	// where it tries to get the etcdadm cluster for the first time before it generates the resources.
	if cluster.Spec.ExternalEtcdConfiguration != nil && etcdadmCluster == nil {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.ExternalEtcdNotAvailable, clusterv1.ConditionSeverityInfo, "Etcd cluster is not available")
		return
	}
	// Make sure etcd machine is ready before marking ControlPlaneReady status to true
	if cluster.Spec.ExternalEtcdConfiguration != nil && !etcdadmClusterReady(etcdadmCluster) {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.RollingUpgradeInProgress, clusterv1.ConditionSeverityInfo, "Etcd is not ready")
		return
	}
	updateControlPlaneReadyCondition(cluster, kcp)
}

// updateControlPlaneReadyCondition updates the ControlPlaneReady condition, after checking the state of the control plane
// in the cluster.
func updateControlPlaneReadyCondition(cluster *anywherev1.Cluster, kcp *controlplanev1.KubeadmControlPlane) {
	initializedCondition := conditions.Get(cluster, anywherev1.ControlPlaneInitializedCondition)
	if initializedCondition.Status != "True" {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, initializedCondition.Reason, initializedCondition.Severity, initializedCondition.Message)
		return
	}

	if kcp == nil {
		return
	}

	// We make sure to check that the status is up to date before using it
	if kcp.Status.ObservedGeneration != kcp.ObjectMeta.Generation {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.OutdatedInformationReason, clusterv1.ConditionSeverityInfo, "")
		return
	}

	// The control plane should be marked ready when the count specified in the spec is
	// equal to the ready number of nodes in the cluster and they're all of the right version specified.

	expected := cluster.Spec.ControlPlaneConfiguration.Count
	totalReplicas := int(kcp.Status.Replicas)

	// First, in the case of a rolling upgrade, we get the number of outdated nodes, and as long as there are some,
	// we want to reflect in the message that the Cluster is in progres upgdating the old nodes with the
	// the new machine spec.
	updatedReplicas := int(kcp.Status.UpdatedReplicas)
	totalOutdated := totalReplicas - updatedReplicas

	if totalOutdated > 0 {
		upgradeReason := anywherev1.RollingUpgradeInProgress
		if cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
			if cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.Type == anywherev1.InPlaceStrategyType {
				upgradeReason = anywherev1.InPlaceUpgradeInProgress
			}
		}
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, upgradeReason, clusterv1.ConditionSeverityInfo, "Control plane nodes not up-to-date yet, %d upgrading (%d up to date)", totalReplicas, updatedReplicas)
		return
	}

	// Then, we check that the number of nodes in the cluster match the expected amount. If not, we
	// mark that the Cluster is scaling up or scale down the control plane replicas to the expected amount.
	if totalReplicas < expected {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up control plane nodes, %d expected (%d actual)", expected, totalReplicas)
		return
	}

	if totalReplicas > expected {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.ScalingDownReason, clusterv1.ConditionSeverityInfo, "Scaling down control plane nodes, %d expected (%d actual)", expected, totalReplicas)
		return
	}

	readyReplicas := int(kcp.Status.ReadyReplicas)
	if readyReplicas != expected {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.NodesNotReadyReason, clusterv1.ConditionSeverityInfo, "Control plane nodes not ready yet, %d expected (%d ready)", expected, readyReplicas)
		return
	}

	// We check the condition signifying the overall health of the control plane components. Usually, the control plane should be healthy
	// at this point but if that is not the case, we report it as an error.
	kcpControlPlaneHealthyCondition := conditions.Get(kcp, controlplanev1.ControlPlaneComponentsHealthyCondition)
	if kcpControlPlaneHealthyCondition != nil && kcpControlPlaneHealthyCondition.Status == v1.ConditionFalse {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.ControlPlaneComponentsUnhealthyReason, clusterv1.ConditionSeverityError, kcpControlPlaneHealthyCondition.Message)
		return
	}

	conditions.MarkTrue(cluster, anywherev1.ControlPlaneReadyCondition)
}

// updateControlPlaneInitializedCondition updates the ControlPlaneInitialized condition if it hasn't already been set.
// This condition should be set only once.
func updateControlPlaneInitializedCondition(cluster *anywherev1.Cluster, kcp *controlplanev1.KubeadmControlPlane) {
	// Return early if the ControlPlaneInitializedCondition is already "True"
	if conditions.IsTrue(cluster, anywherev1.ControlPlaneInitializedCondition) {
		return
	}

	if kcp == nil {
		conditions.Set(cluster, controlPlaneInitializationInProgressCondition())
		return
	}

	// We make sure to check that the status is up to date before using it
	if kcp.Status.ObservedGeneration != kcp.ObjectMeta.Generation {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneInitializedCondition, anywherev1.OutdatedInformationReason, clusterv1.ConditionSeverityInfo, "")
		return
	}

	// Then, we'll check explicitly for that the control plane is available. This way, we do not rely on CAPI
	// to implicitly to fill out our conditions reasons, and we can have custom messages.
	available := conditions.IsTrue(kcp, controlplanev1.AvailableCondition)
	if !available {
		conditions.Set(cluster, controlPlaneInitializationInProgressCondition())
		return
	}

	conditions.MarkTrue(cluster, anywherev1.ControlPlaneInitializedCondition)
}

// updateWorkersReadyCondition updates the WorkersReadyCondition condition after checking the state of the worker node groups
// in the cluster.
func updateWorkersReadyCondition(cluster *anywherev1.Cluster, machineDeployments []clusterv1.MachineDeployment) {
	initializedCondition := conditions.Get(cluster, anywherev1.ControlPlaneInitializedCondition)
	if initializedCondition.Status != "True" {
		conditions.MarkFalse(cluster, anywherev1.WorkersReadyCondition, anywherev1.ControlPlaneNotInitializedReason, clusterv1.ConditionSeverityInfo, "")
		return
	}

	totalExpected := 0
	for _, wng := range cluster.Spec.WorkerNodeGroupConfigurations {
		totalExpected += *wng.Count
	}

	// First, we need aggregate the number of nodes across worker node groups  to be able to assess the condition of the workers
	// as a whole.
	totalReadyReplicas := 0
	totalUpdatedReplicas := 0
	totalReplicas := 0

	for _, md := range machineDeployments {
		// We make sure to check that the status is up to date before using the information from the machine deployment status.
		if md.Status.ObservedGeneration != md.ObjectMeta.Generation {
			conditions.MarkFalse(cluster, anywherev1.WorkersReadyCondition, anywherev1.OutdatedInformationReason, clusterv1.ConditionSeverityInfo, "Worker node group %s status not up to date yet", md.Name)
			return
		}

		totalReadyReplicas += int(md.Status.ReadyReplicas)
		totalUpdatedReplicas += int(md.Status.UpdatedReplicas)
		totalReplicas += int(md.Status.Replicas)
	}

	// There may be worker nodes that are not up to date yet in the case of a rolling upgrade,
	// so reflect that on the conditon with an appropriate message.
	totalOutdated := totalReplicas - totalUpdatedReplicas
	if totalOutdated > 0 {
		upgradeReason := anywherev1.RollingUpgradeInProgress
		// We are checking the control plane configuration here because we already validate that all the machines
		// have the same upgrade strategy.
		if cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
			if cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.Type == anywherev1.InPlaceStrategyType {
				upgradeReason = anywherev1.InPlaceUpgradeInProgress
			}
		}
		conditions.MarkFalse(cluster, anywherev1.WorkersReadyCondition, upgradeReason, clusterv1.ConditionSeverityInfo, "Worker nodes not up-to-date yet, %d upgrading (%d up to date)", totalReplicas, totalUpdatedReplicas)
		return
	}

	// If the number of worker nodes replicas need to be scaled up.
	if totalReplicas < totalExpected {
		conditions.MarkFalse(cluster, anywherev1.WorkersReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, %d expected (%d actual)", totalExpected, totalReplicas)
		return
	}

	// If the number of worker nodes replicas need to be scaled down.
	if totalReplicas > totalExpected {
		conditions.MarkFalse(cluster, anywherev1.WorkersReadyCondition, anywherev1.ScalingDownReason, clusterv1.ConditionSeverityInfo, "Scaling down worker nodes, %d expected (%d actual)", totalExpected, totalReplicas)
		return
	}

	if totalReadyReplicas != totalExpected {
		conditions.MarkFalse(cluster, anywherev1.WorkersReadyCondition, anywherev1.NodesNotReadyReason, clusterv1.ConditionSeverityInfo, "Worker nodes not ready yet, %d expected (%d ready)", totalExpected, totalReadyReplicas)
		return
	}

	conditions.MarkTrue(cluster, anywherev1.WorkersReadyCondition)
}

// controlPlaneInitializationInProgressCondition returns a new "False" condition for the ControlPlaneInitializationInProgress reason.
func controlPlaneInitializationInProgressCondition() *anywherev1.Condition {
	return conditions.FalseCondition(anywherev1.ControlPlaneInitializedCondition, anywherev1.ControlPlaneInitializationInProgressReason, clusterv1.ConditionSeverityInfo, "The first control plane instance is not available yet")
}
