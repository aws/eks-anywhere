package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// NodeUpgradeKind stores the Kind for NodeUpgrade.
	NodeUpgradeKind = "NodeUpgrade"

	// UpgraderPodCreated reports whether the upgrader pod has been created for the node upgrade.
	UpgraderPodCreated ConditionType = "UpgraderPodCreated"

	// BinariesCopied reports whether the binaries have been copied over by the component copier container.
	BinariesCopied ConditionType = "BinariesCopied"

	// ContainerdUpgraded reports whether containerd has been upgraded.
	ContainerdUpgraded ConditionType = "ContainerdUpgraded"

	// CNIPluginsUpgraded reports whether the CNI plugins has been upgraded.
	CNIPluginsUpgraded ConditionType = "CNIPluginsUpgraded"

	// KubeadmUpgraded reports whether Kubeadm has been upgraded.
	KubeadmUpgraded ConditionType = "KubeadmUpgraded"

	// KubeletUpgraded reports whether kubelet has been upgraded.
	KubeletUpgraded ConditionType = "KubeletUpgraded"

	// PostUpgradeCleanupCompleted reports whether the post upgrade operations have been completed.
	PostUpgradeCleanupCompleted ConditionType = "PostUpgradeCleanupCompleted"
)

// NodeUpgradeSpec defines the desired state of NodeUpgrade.
type NodeUpgradeSpec struct {
	// Machine is a reference to the CAPI Machine that needs to be upgraded.
	Machine corev1.ObjectReference `json:"machine"`

	// KubernetesVersion refers to the Kubernetes version to upgrade the node to.
	KubernetesVersion string `json:"kubernetesVersion"`

	// EtcdVersion refers to the version of ETCD to upgrade to.
	// This field is optional and only gets used for control plane nodes.
	// +optional
	EtcdVersion *string `json:"etcdVersion,omitempty"`

	// FirstNodeToBeUpgraded signifies that the Node is the first node to be upgraded.
	// This flag is only valid for control plane nodes and ignored for worker nodes.
	// +optional
	FirstNodeToBeUpgraded bool `json:"firstNodeToBeUpgraded,omitempty"`
}

// NodeUpgradeStatus defines the observed state of NodeUpgrade.
type NodeUpgradeStatus struct {
	// Conditions defines current state of the NodeUpgrade,
	// including the state of init containers, that facilitate the upgrade.
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`

	// Completed denotes that the upgrader has completed running all the operations
	// and the node is successfully upgraded.
	// +optional
	Completed bool `json:"completed"`

	// ObservedGeneration is the latest generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:path=nodeupgrades,shortName=nu,scope=Namespaced,singular=nodeupgrade
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".spec.machine.name",description="Machine"
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.completed",description="Denotes whether the upgrade has finished or not"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of Control Plane Upgrade"
//+kubebuilder:printcolumn:name="KubernetesVersion",type="string",JSONPath=".spec.kubernetesVersion",description="Requested Kubernetes version"

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

// GetConditions returns all the Conditions for the NodeUpgrade object.
func (n *NodeUpgrade) GetConditions() clusterv1.Conditions {
	return n.Status.Conditions
}

// SetConditions sets the Conditons on the NodeUpgrade object.
func (n *NodeUpgrade) SetConditions(conditions clusterv1.Conditions) {
	n.Status.Conditions = conditions
}
