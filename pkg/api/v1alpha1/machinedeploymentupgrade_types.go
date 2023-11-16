package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineDeploymentUpgradeKind stores the Kind for MachineDeploymentUpgrade.
const MachineDeploymentUpgradeKind = "MachineDeploymentUpgrade"

// MachineDeploymentUpgradeSpec defines the desired state of MachineDeploymentUpgrade.
type MachineDeploymentUpgradeSpec struct {
	Cluster                Ref    `json:"cluster"`
	MachineDeployment      Ref    `json:"controlPlane"`
	MachinesRequireUpgrade []Ref  `json:"machinesRequireUpgrade"`
	KubernetesVersion      string `json:"kubernetesVersion"`
	KubeletVersion         string `json:"kubeletVersion"`
	KubeadmClusterConfig   string `json:"kubeadmClusterConfig"`
}

// MachineDeploymentUpgradeStatus defines the observed state of MachineDeploymentUpgrade.
type MachineDeploymentUpgradeStatus struct {
	RequireUpgrade int64 `json:"requireUpgrade"`
	Upgraded       int64 `json:"upgraded"`
	Ready          bool  `json:"ready"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MachineDeploymentUpgrade is the Schema for the machinedeploymentupgrades API.
type MachineDeploymentUpgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineDeploymentUpgradeSpec   `json:"spec,omitempty"`
	Status MachineDeploymentUpgradeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineDeploymentUpgradeList contains a list of MachineDeploymentUpgrade.
type MachineDeploymentUpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineDeploymentUpgrade `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachineDeploymentUpgrade{}, &MachineDeploymentUpgradeList{})
}
