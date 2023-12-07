package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ControlPlaneUpgradeKind stores the kind for ControlPlaneUpgrade.
const ControlPlaneUpgradeKind = "ControlPlaneUpgrade"

// ControlPlaneUpgradeSpec defines the desired state of ControlPlaneUpgrade.
type ControlPlaneUpgradeSpec struct {
	Cluster                Ref     `json:"cluster"`
	ControlPlane           Ref     `json:"controlPlane"`
	MachinesRequireUpgrade []Ref   `json:"machinesRequireUpgrade"`
	KubernetesVersion      string  `json:"kubernetesVersion"`
	KubeletVersion         string  `json:"kubeletVersion"`
	EtcdVersion            *string `json:"etcdVersion,omitempty"`
	CoreDNSVersion         *string `json:"coreDNSVersion,omitempty"`
	KubeadmClusterConfig   string  `json:"kubeadmClusterConfig"`
}

// ControlPlaneUpgradeStatus defines the observed state of ControlPlaneUpgrade.
type ControlPlaneUpgradeStatus struct {
	RequireUpgrade int64 `json:"requireUpgrade,omitempty"`
	Upgraded       int64 `json:"upgraded,omitempty"`
	Ready          bool  `json:"ready,omitempty"`
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
