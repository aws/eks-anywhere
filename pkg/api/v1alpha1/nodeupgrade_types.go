package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// NodeUpgradeKind stores the Kind for NodeUpgrade.
const NodeUpgradeKind = "NodeUpgrade"

// NodeUpgradeSpec defines the desired state of NodeUpgrade.
type NodeUpgradeSpec struct {
	Machine           Ref     `json:"machine"`
	Node              Ref     `json:"node"`
	KubernetesVersion string  `json:"kubernetesVersion"`
	KubeletVersion    string  `json:"kubeletVersion"`
	EtcdVersion       *string `json:"etcdVersion,omitempty"`
	CoreDNSVersion    *string `json:"coreDNSVersion,omitempty"`
}

// NodeUpgradeStatus defines the observed state of NodeUpgrade.
type NodeUpgradeStatus struct {
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
	Phase      string               `json:"phase"`
	Completed  bool                 `json:"completed"`
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
