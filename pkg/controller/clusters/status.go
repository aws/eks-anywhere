package clusters

import (
	"fmt"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

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

	kcpCondition, err := checkKubeadmControlPlaneError(cluster, anywherev1.ControlPlaneInitializedCondition, kcp, controlplanev1.AvailableCondition)
	if err != nil {
		return err
	}

	if kcpCondition != nil {
		conditions.MarkFalse(cluster, anywherev1.ControlPlaneInitializedCondition, kcpCondition.Reason, kcpCondition.Severity, kcpCondition.Message)
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
