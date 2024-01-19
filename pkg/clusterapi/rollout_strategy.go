package clusterapi

import (
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// SetUpgradeRolloutStrategyInKubeadmControlPlane updates the kubeadm control plane with the upgrade rollout strategy defined in an eksa cluster.
func SetUpgradeRolloutStrategyInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, rolloutStrategy *anywherev1.ControlPlaneUpgradeRolloutStrategy) {
	if rolloutStrategy != nil {
		maxSurge := intstr.FromInt(rolloutStrategy.RollingUpdate.MaxSurge)
		kcp.Spec.RolloutStrategy = &controlplanev1.RolloutStrategy{
			Type: controlplanev1.RollingUpdateStrategyType,
			RollingUpdate: &controlplanev1.RollingUpdate{
				MaxSurge: &maxSurge,
			},
		}
	}
}

// SetUpgradeRolloutStrategyInMachineDeployment updates the machine deployment with the upgrade rollout strategy defined in an eksa cluster.
func SetUpgradeRolloutStrategyInMachineDeployment(md *clusterv1.MachineDeployment, rolloutStrategy *anywherev1.WorkerNodesUpgradeRolloutStrategy) {
	if rolloutStrategy != nil {
		maxSurge := intstr.FromInt(rolloutStrategy.RollingUpdate.MaxSurge)
		maxUnavailable := intstr.FromInt(rolloutStrategy.RollingUpdate.MaxUnavailable)
		md.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
			Type: clusterv1.RollingUpdateMachineDeploymentStrategyType,
			RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
				MaxSurge:       &maxSurge,
				MaxUnavailable: &maxUnavailable,
			},
		}
	}
}
