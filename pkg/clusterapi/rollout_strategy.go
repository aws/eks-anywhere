package clusterapi

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// SetUpgradeRolloutStrategyInKubeadmControlPlane updates the kubeadm control plane with the upgrade rollout strategy defined in an eksa cluster.
func SetUpgradeRolloutStrategyInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, rolloutStrategy *anywherev1.ControlPlaneUpgradeRolloutStrategy) {
	if rolloutStrategy != nil {
		maxSurge := intstr.FromInt(rolloutStrategy.RollingUpdate.MaxSurge)
		kcp.Spec.Rollout.Strategy = controlplanev1beta2.KubeadmControlPlaneRolloutStrategy{
			Type: controlplanev1beta2.RollingUpdateStrategyType,
			RollingUpdate: controlplanev1beta2.KubeadmControlPlaneRolloutStrategyRollingUpdate{
				MaxSurge: &maxSurge,
			},
		}
	}
}

// SetUpgradeRolloutStrategyInMachineDeployment updates the machine deployment with the upgrade rollout strategy defined in an eksa cluster.
func SetUpgradeRolloutStrategyInMachineDeployment(md *clusterv1beta2.MachineDeployment, rolloutStrategy *anywherev1.WorkerNodesUpgradeRolloutStrategy) {
	if rolloutStrategy != nil {
		maxSurge := intstr.FromInt(rolloutStrategy.RollingUpdate.MaxSurge)
		maxUnavailable := intstr.FromInt(rolloutStrategy.RollingUpdate.MaxUnavailable)
		md.Spec.Rollout = clusterv1beta2.MachineDeploymentRolloutSpec{
			Strategy: clusterv1beta2.MachineDeploymentRolloutStrategy{
				Type: clusterv1beta2.RollingUpdateMachineDeploymentStrategyType,
				RollingUpdate: clusterv1beta2.MachineDeploymentRolloutStrategyRollingUpdate{
					MaxSurge:       &maxSurge,
					MaxUnavailable: &maxUnavailable,
				},
			},
		}
	}
}
