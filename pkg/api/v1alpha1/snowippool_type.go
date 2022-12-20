package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

// SnowIPPoolSpec defines the desired state of SnowIPPool.
type SnowIPPoolSpec struct {
	// IPPools defines a list of ip pool for the DNI.
	Pools []snowv1.IPPool `json:"pools,omitempty"`
}

// SnowIPPoolStatus defines the observed state of SnowIPPool.
type SnowIPPoolStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SnowIPPool is the Schema for the SnowIPPools API.
type SnowIPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnowIPPoolSpec   `json:"spec,omitempty"`
	Status SnowIPPoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SnowIPPoolList contains a list of SnowIPPool.
type SnowIPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowIPPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SnowIPPool{}, &SnowIPPoolList{})
}
