package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineDeploymentUpgradeKind stores the Kind for MachineDeploymentUpgrade.
const MachineDeploymentUpgradeKind = "MachineDeploymentUpgrade"

// MachineDeploymentUpgradeSpec defines the desired state of MachineDeploymentUpgrade.
type MachineDeploymentUpgradeSpec struct {
	// MachineDeployment is a reference to the KubeadmControlPlane object to upgrade.
	MachineDeployment corev1.ObjectReference `json:"machineDeployment"`

	// MachinesRequireUpgrade is a list of references to CAPI machines that need to be upgraded.
	MachinesRequireUpgrade []corev1.ObjectReference `json:"machinesRequireUpgrade"`

	// KubernetesVersion refers to the Kubernetes version to upgrade the control planes to.
	KubernetesVersion string `json:"kubernetesVersion"`

	// MachineSpecData is a base64 encoded json string value of the machineDeplopyment.Spec.Template.Spec field that's specification of the desired behavior of the machine.
	MachineSpecData string `json:"machineSpecData"`
}

// MachineDeploymentUpgradeStatus defines the observed state of MachineDeploymentUpgrade.
type MachineDeploymentUpgradeStatus struct {
	// RequireUpgrade is the number of machines in the MachineDeployment that still need to be upgraded.
	RequireUpgrade int64 `json:"requireUpgrade,omitempty"`

	// Upgraded is the number of machines in the MachineDeployment that have been upgraded.
	Upgraded int64 `json:"upgraded,omitempty"`

	// Ready denotes that the all machines in the MachineDeployment have finished upgrading and are ready.
	Ready bool `json:"ready,omitempty"`
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
