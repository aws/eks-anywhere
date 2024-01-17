package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ControlPlaneUpgradeKind stores the kind for ControlPlaneUpgrade.
const ControlPlaneUpgradeKind = "ControlPlaneUpgrade"

// ControlPlaneUpgradeSpec defines the desired state of ControlPlaneUpgrade.
type ControlPlaneUpgradeSpec struct {
	// ControlPlane is a reference to the KubeadmControlPlane object to upgrade.
	ControlPlane corev1.ObjectReference `json:"controlPlane"`

	// MachinesRequireUpgrade is a list of references to CAPI machines that need to be upgraded.
	MachinesRequireUpgrade []corev1.ObjectReference `json:"machinesRequireUpgrade"`

	// KubernetesVersion refers to the Kubernetes version to upgrade the control planes to.
	KubernetesVersion string `json:"kubernetesVersion"`

	// EtcdVersion refers to the version of ETCD to upgrade to.
	EtcdVersion string `json:"etcdVersion"`
}

// ControlPlaneUpgradeStatus defines the observed state of ControlPlaneUpgrade.
type ControlPlaneUpgradeStatus struct {
	// RequireUpgrade is the number of machines that still need to be upgraded.
	RequireUpgrade int64 `json:"requireUpgrade,omitempty"`

	// Upgraded is the number of machines that have been upgraded.
	Upgraded int64 `json:"upgraded,omitempty"`

	// Ready denotes that the all control planes have finished upgrading and are ready.
	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ControlPlaneUpgrade is the Schema for the controlplaneupgrade API.
type ControlPlaneUpgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneUpgradeSpec   `json:"spec,omitempty"`
	Status ControlPlaneUpgradeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ControlPlaneUpgradeList contains a list of ControlPlaneUpgradeSpec.
type ControlPlaneUpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlaneUpgrade `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlaneUpgrade{}, &ControlPlaneUpgradeList{})
}
