package clusters

import (
	"context"

	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

// UpdateClusterStatusForControlPlane checks the current state of the Cluster's control plane and updates the
// Cluster status information.
func UpdateClusterStatusForControlPlane(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) error {
	kcp, err := controller.GetKubeadmControlPlane(ctx, client, cluster)
	if err != nil {
		return errors.Wrapf(err, "getting kubeadmcontrolplane")
	}

	updateControlPlaneInitializedCondition(cluster, kcp)

	return nil
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

// controlPlaneInitializationInProgressCondition returns a new "False" condition for the ControlPlaneInitializationInProgress reason.
func controlPlaneInitializationInProgressCondition() *anywherev1.Condition {
	return conditions.FalseCondition(anywherev1.ControlPlaneInitializedCondition, anywherev1.ControlPlaneInitializationInProgressReason, clusterv1.ConditionSeverityInfo, "The first control plane instance is not available yet")
}
