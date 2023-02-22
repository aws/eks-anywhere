package tinkerbell

import (
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

// VersionComparator enables kubernetes version retrieval.
type VersionComparator interface {
	GetVersion() string
}

// ReplicaComparator enables replica number retrieval.
type ReplicaComparator interface {
	GetReplicas() int
}

// KubeadmControlPlaneComparator wraps the KubeadmControlPlane to enable Comparator functions.
type KubeadmControlPlaneComparator struct {
	*controlplanev1.KubeadmControlPlane
}

// GetVersion retrieves KubeadmControlPlane replicas.
func (k *KubeadmControlPlaneComparator) GetVersion() string {
	return k.Spec.Version
}

// GetReplicas retrieves KubeadmControlPlane replicas.
func (k *KubeadmControlPlaneComparator) GetReplicas() int {
	return int(*k.Spec.Replicas)
}

// MachineDeploymentComparator wraps the MachineDeployment to enable Comparator functions.
type MachineDeploymentComparator struct {
	*clusterv1.MachineDeployment
}

// GetReplicas retrieves MachineDeployment replicas.
func (m *MachineDeploymentComparator) GetReplicas() int {
	return int(*m.Spec.Replicas)
}

// KubernetesVersionChange detects kubernetes version has changed between two Comparator objects.
func KubernetesVersionChange(oldObj VersionComparator, newObj VersionComparator) bool {
	return oldObj.GetVersion() != newObj.GetVersion()
}

// ReplicasChange returns the difference in replica count between new and old Comparator objects.
func ReplicasChange(oldObj ReplicaComparator, newObj ReplicaComparator) int {
	return newObj.GetReplicas() - oldObj.GetReplicas()
}
