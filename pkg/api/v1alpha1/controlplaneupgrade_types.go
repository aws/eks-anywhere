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

	// ControlPlaneSpecData contains base64 encoded KCP spec that's used to update
	// the statuses of CAPI objects once the control plane upgrade is done.
	// This field is needed so that we have a static copy of the control plane spec
	// in case it gets modified after the ControlPlaneUpgrade was created,
	// as ControlPlane is a reference to the object in real time.
	ControlPlaneSpecData string `json:"controlPlaneSpecData"`
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
//+kubebuilder:resource:path=controlplaneupgrades,shortName=cpu,scope=Namespaced,singular=controlplaneupgrade
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="KubeadmControlPlane",type="string",JSONPath=".spec.controlPlane.name",description="KubeadmControlPlane"
//+kubebuilder:printcolumn:name="Upgraded",type="string",JSONPath=".status.upgraded",description="Control Plane machines that are already upgraded"
//+kubebuilder:printcolumn:name="PendingUpgrade",type="string",JSONPath=".status.requireUpgrade",description="Control Plane machines that still require upgrade"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Denotes whether the upgrade has finished or not"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of Control Plane Upgrade"
//+kubebuilder:printcolumn:name="KubernetesVersion",type="string",JSONPath=".spec.kubernetesVersion",description="Requested Kubernetes version"

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
