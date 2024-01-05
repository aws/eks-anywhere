package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ControlPlaneUpgradeKind stores the kind for ControlPlaneUpgrade.
const ControlPlaneUpgradeKind = "ControlPlaneUpgrade"

// ControlPlaneUpgradeSpec defines the desired state of ControlPlaneUpgrade.
type ControlPlaneUpgradeSpec struct {
	ControlPlane corev1.ObjectReference `json:"controlPlane"`
	EtcdVersion  string                 `json:"etcdVersion"`
}

// ControlPlaneUpgradeStatus defines the observed state of ControlPlaneUpgrade.
type ControlPlaneUpgradeStatus struct {
	RequireUpgrade int64          `json:"requireUpgrade,omitempty"`
	Upgraded       int64          `json:"upgraded,omitempty"`
	Ready          bool           `json:"ready,omitempty"`
	MachineState   []MachineState `json:"machineState,omitempty"`
}

// MachineState stores the name of machine and whether it has been upgraded or not.
type MachineState struct {
	Name     string `json:"name"`
	Upgraded bool   `json:"upgraded"`
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
