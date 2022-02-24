package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SnowDatacenterConfigSpec defines the desired state of SnowDatacenterConfig
type SnowDatacenterConfigSpec struct { // Important: Run "make generate" to regenerate code after modifying this file
}

// SnowDatacenterConfigStatus defines the observed state of SnowDatacenterConfig
type SnowDatacenterConfigStatus struct{}

// SnowDatacenterConfig is the Schema for the SnowDatacenterConfigs API
type SnowDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnowDatacenterConfigSpec   `json:"spec,omitempty"`
	Status SnowDatacenterConfigStatus `json:"status,omitempty"`
}

func (s *SnowDatacenterConfig) Kind() string {
	return s.TypeMeta.Kind
}

func (s *SnowDatacenterConfig) ExpectedKind() string {
	return VSphereDatacenterKind
}

// +kubebuilder:object:generate=false

// Same as SnowDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type SnowDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec SnowDatacenterConfigSpec `json:"spec,omitempty"`
}
