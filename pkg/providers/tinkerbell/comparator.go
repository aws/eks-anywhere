package tinkerbell

import (
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

// Versionable enables kubernetes version retrieval.
type Versionable interface {
	Version() string
}

// Replicable enables replica number retrieval.
type Replicable interface {
	Replicas() int
}

// KubeadmControlPlane wraps the KubeadmControlPlane to enable Comparator functions.
type KubeadmControlPlane struct {
	*controlplanev1.KubeadmControlPlane
}

// Version retrieves KubeadmControlPlane replicas.
func (k *KubeadmControlPlane) Version() string {
	return k.Spec.Version
}

// Replicas retrieves KubeadmControlPlane replicas.
func (k *KubeadmControlPlane) Replicas() int {
	return int(*k.Spec.Replicas)
}

// MachineDeployment implements Replicable by decorating a MachineDeployment.
type MachineDeployment struct {
	*clusterv1.MachineDeployment
}

// Replicas retrieves MachineDeployment replicas.
func (m *MachineDeployment) Replicas() int {
	return int(*m.Spec.Replicas)
}

// HasKubernetesVersionChange detects kubernetes version has changed between two Comparator objects.
func HasKubernetesVersionChange(current Versionable, desired Versionable) bool {
	return current.Version() != desired.Version()
}

// ReplicasDiff returns the difference in replica count between new and old Comparator objects.
func ReplicasDiff(current Replicable, desired Replicable) int {
	return desired.Replicas() - current.Replicas()
}
