package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VSphereDatacenterConfigSpec defines the desired state of VSphereDatacenterConfig
type VSphereDatacenterConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	Datacenter string `json:"datacenter"`
	Network    string `json:"network"`
	Server     string `json:"server"`
	Thumbprint string `json:"thumbprint"`
	Insecure   bool   `json:"insecure"`
}

// VSphereDatacenterConfigStatus defines the observed state of VSphereDatacenterConfig
type VSphereDatacenterConfigStatus struct { // Important: Run "make generate" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VSphereDatacenterConfig is the Schema for the VSphereDatacenterConfigs API
type VSphereDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VSphereDatacenterConfigSpec   `json:"spec,omitempty"`
	Status VSphereDatacenterConfigStatus `json:"status,omitempty"`
}

func (v *VSphereDatacenterConfig) Kind() string {
	return v.TypeMeta.Kind
}

func (v *VSphereDatacenterConfig) ExpectedKind() string {
	return VSphereDatacenterKind
}

func (v *VSphereDatacenterConfig) PauseReconcile() {
	if v.Annotations == nil {
		v.Annotations = map[string]string{}
	}
	v.Annotations[pausedAnnotation] = "true"
}

func (v *VSphereDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := v.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (v *VSphereDatacenterConfig) ClearPauseAnnotation() {
	if v.Annotations != nil {
		delete(v.Annotations, pausedAnnotation)
	}
}

// +kubebuilder:object:generate=false

// Same as VSphereDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type VSphereDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec VSphereDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// VSphereDatacenterConfigList contains a list of VSphereDatacenterConfig
type VSphereDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VSphereDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VSphereDatacenterConfig{}, &VSphereDatacenterConfigList{})
}
