package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// NodeUpgradeKind stores the Kind for NodeUpgrade.
	NodeUpgradeKind = "NodeUpgrade"

	UpgraderPodCreated = "UpgraderPodCreated"

	//
	BinariesCopied ConditionType = "BinariesCopied"

	ContainerdUpgraded ConditionType = "ContainerdUpgraded"

	CNIPluginsUpgraded ConditionType = "CNIPluginsUpgraded"

	KubeadmUpgraded ConditionType = "KubeadmUpgraded"

	KubeletUpgraded ConditionType = "KubeletUpgraded"

	PostUpgradeCleanupCompleted ConditionType = "PostUpgradeCleanupCompleted"
)

// NodeUpgradeSpec defines the desired state of NodeUpgrade.
type NodeUpgradeSpec struct {
	// Machine is a reference to the CAPI Machine that needs to be upgraded.
	Machine           corev1.ObjectReference `json:"machine"`
	KubernetesVersion string                 `json:"kubernetesVersion"`
	KubeletVersion    string                 `json:"kubeletVersion"`
	EtcdVersion       *string                `json:"etcdVersion,omitempty"`
	CoreDNSVersion    *string                `json:"coreDNSVersion,omitempty"`
}

// NodeUpgradeStatus defines the observed state of NodeUpgrade.
type NodeUpgradeStatus struct {
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
	// +optional
	Completed bool `json:"completed,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NodeUpgrade is the Schema for the nodeupgrades API.
type NodeUpgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeUpgradeSpec   `json:"spec,omitempty"`
	Status NodeUpgradeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeUpgradeList contains a list of NodeUpgrade.
type NodeUpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeUpgrade `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeUpgrade{}, &NodeUpgradeList{})
}

func (n *NodeUpgrade) GetConditions() clusterv1.Conditions {
	return n.Status.Conditions
}

func (n *NodeUpgrade) SetConditions(conditions clusterv1.Conditions) {
	n.Status.Conditions = conditions
}
