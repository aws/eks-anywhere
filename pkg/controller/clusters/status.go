package clusters

import (
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func checkKubeadmControlPlaneError(cluster *anywherev1.Cluster, clusterConditionType anywherev1.ConditionType, kcp *controlplanev1.KubeadmControlPlane, kcpConditionType clusterv1.ConditionType) (*anywherev1.Condition, error) {
	kcpCondition := conditions.Get(kcp, kcpConditionType)

	if kcpCondition == nil {
		return nil, fmt.Errorf("unable to read condition %s from kubeadmcontrolplane %s", kcpConditionType, kcp.Name)
	}

	// Surface any errors from the KubeadmControlPlane
	if kcpCondition.Severity == clusterv1.ConditionSeverityError {
		return conditions.FalseCondition(clusterConditionType, kcpCondition.Reason, kcpCondition.Severity, kcpCondition.Message), nil
	}

	return nil, nil
}

// UpdateControlPlaneInitializedCondition updates the ControlPlaneInitialized condition if it hasn't already been set.
// This condition should be set only once.
func UpdateControlPlaneInitializedCondition(cluster *anywherev1.Cluster, kcp *controlplanev1.KubeadmControlPlane) error {
	// Return early if the ControlPlaneInitializedCondition is already "True"
	if conditions.IsTrue(cluster, anywherev1.ControlPlaneInitializedCondition) {
		return nil
	}

	if kcp == nil {
		conditions.Set(cluster, waitingForCPInitializedCondition())
		return nil
	}

	con, err := checkKubeadmControlPlaneError(cluster, anywherev1.ControlPlaneInitializedCondition, kcp, controlplanev1.AvailableCondition)
	if err != nil {
		return err
	}

	if con != nil {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneInitializedCondition, con.Reason, con.Severity, con.Message)
		return nil
	}

	available := conditions.IsTrue(kcp, controlplanev1.AvailableCondition)
	if !available {
		conditions.Set(cluster, waitingForCPInitializedCondition())
		return nil
	}

	conditions.MarkTrue(cluster, anywherev1.ControlPlaneInitializedCondition)
	return nil
}

func waitingForCPInitializedCondition() *anywherev1.Condition {
	return conditions.FalseCondition(anywherev1.ControlPlaneInitializedCondition, anywherev1.WaitingForControlPlaneInitializedReason, clusterv1.ConditionSeverityInfo, anywherev1.FirstControlPlaneUnavailableMessage)
}

// UpdateControlPlaneReadyCondition updates the ControlPlaneReady condition. It checks whether the KubeadmControlPlane is ready
// and number of ready vs expected replicas are accounted for.
func UpdateControlPlaneReadyCondition(cluster *anywherev1.Cluster, kcp *controlplanev1.KubeadmControlPlane) error {
	initializedCondition := conditions.Get(cluster, anywherev1.ControlPlaneInitializedCondition)
	if initializedCondition.Status != "True" {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, initializedCondition.Reason, initializedCondition.Severity, initializedCondition.Message)
		return nil
	}

	if kcp == nil {
		return fmt.Errorf("expected reference to kubeadmcontrolplane, but got nil instead ")
	}

	statusUpToDate := kcp.Status.ObservedGeneration == kcp.ObjectMeta.Generation
	// We make sure to check that the status is up to date before using it
	if !statusUpToDate {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.PendingUpdateReason, clusterv1.ConditionSeverityInfo, "")
		return nil
	}

	con, err := checkKubeadmControlPlaneError(cluster, anywherev1.ControlPlaneReadyCondition, kcp, clusterv1.ReadyCondition)
	if err != nil {
		return err
	}

	if con != nil {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, con.Reason, con.Severity, con.Message)
		return nil
	}

	// The control plane should be marked ready when the total number of nodes specified in the spec is
	// equal to the ready amount in the cluster. We need to check this in two parts, since
	// in the case of a rolling upgrade, there can be ready machines in the cluster with the old spec.

	expected := cluster.Spec.ControlPlaneConfiguration.Count
	readyReplicas := int(kcp.Status.ReadyReplicas)
	updatedReplicas := int(kcp.Status.UpdatedReplicas)
	totalReplicas := int(kcp.Status.Replicas)

	// Get the number of outdated nodes, as long as there are some, we want to reflect that in the message
	totalOutdated := totalReplicas - updatedReplicas
	if totalOutdated > 0 {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.WaitingForControlPlaneReadyReason, clusterv1.ConditionSeverityInfo, "Control plane nodes not up-to-date yet, %d rolling (%d up to date)", expected, updatedReplicas)
		return nil
	}

	// First, we compare the total number of readyReplicas to the expected number,
	if readyReplicas != expected {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneReadyCondition, anywherev1.WaitingForControlPlaneReadyReason, clusterv1.ConditionSeverityInfo, "Control plane nodes not ready yet, %d expected (%d ready)", expected, readyReplicas)
		return nil
	}

	conditions.MarkTrue(cluster, anywherev1.ControlPlaneReadyCondition)
	return nil
}

// UpdateWorkersReadyCondition checks the WorkersReadyConditon condition. It checks whether each MachineDeployment is ready
// and number of ready vs expected replicas are accounted for.
func UpdateWorkersReadyCondition(cluster *anywherev1.Cluster, machineDeployments []clusterv1.MachineDeployment) error {
	expected := 0
	for _, wng := range cluster.Spec.WorkerNodeGroupConfigurations {
		expected += *wng.Count
	}

	// The workers condition should be marked ready when the total number of nodes specified in the spec is
	// equal to the ready amount in the cluster and there are no lingering old nodes. We need to check this in
	// two parts, since in the case of a rolling upgrade, there can be ready machines in the cluster with the old spec.

	totalReady := 0
	totalUpdatedReplicas := 0
	totalReplicas := 0

	for _, md := range machineDeployments {
		// We make sure to check that the status is up to date before using it
		statusUpToDate := md.Status.ObservedGeneration == md.ObjectMeta.Generation
		if !statusUpToDate {
			conditions.MarkFalse(cluster, anywherev1.WorkersReadyConditon, anywherev1.PendingUpdateReason, clusterv1.ConditionSeverityInfo, "Waiting to update machine deployment %s", md.Name)
			return nil
		}

		totalReady += int(md.Status.ReadyReplicas)
		totalUpdatedReplicas += int(md.Status.UpdatedReplicas)
		totalReplicas += int(md.Status.Replicas)
	}

	// First, get the number of outdated nodes, as long as there are some, we want to reflect that in the message
	totalOutdated := totalReplicas - totalUpdatedReplicas
	if totalOutdated > 0 {
		conditions.MarkFalse(cluster, anywherev1.WorkersReadyConditon, anywherev1.WaitingForWorkersReadyReason, clusterv1.ConditionSeverityInfo, "Worker nodes not up-to-date yet, %d rolling (%d up to date)", totalReplicas, totalUpdatedReplicas)
		return nil
	}

	// Then, we compare the total number of readyReplicas to the expected number,
	if totalReady != expected {
		conditions.MarkFalse(cluster, anywherev1.WorkersReadyConditon, anywherev1.WaitingForWorkersReadyReason, clusterv1.ConditionSeverityInfo, "Worker nodes not ready yet, %d expected (%d ready)", expected, totalReady)
		return nil
	}

	conditions.MarkTrue(cluster, anywherev1.WorkersReadyConditon)
	return nil
}

// UpdateSelfManagedClusterDefaultCNIConfiguredCondition is responsible for updating the DefaultCNIConfiguredCondition for self-managed clusters.
// This is meant to be done in the CNI recondile, however, self-managed clusters do not use the it, so this status would never get resolved.
func UpdateSelfManagedClusterDefaultCNIConfiguredCondition(cluster *anywherev1.Cluster) {
	if !conditions.IsTrue(cluster, anywherev1.ControlPlaneReadyCondition) {
		return
	}

	ciliumCfg := cluster.Spec.ClusterNetwork.CNIConfig.Cilium

	// Though it may be installed initially to successfully create the cluster,
	// if the CNI is configured to skip upgrades, we want to mark the condition as "False"
	// as maintenance is not being performed by EKS-A.

	// Otherwise, we can assume the CNI has been
	// configured at this point after confirming the control plane is ready
	if !ciliumCfg.IsManaged() {
		conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, clusterv1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades")
	} else {
		conditions.MarkTrue(cluster, anywherev1.DefaultCNIConfiguredCondition)
	}
}
