package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AWSDatacenterConfigSpec defines the desired state of AWSDatacenterConfig.
type AWSDatacenterConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	Region string `json:"region"`
	AmiID  string `json:"amiID"`
}

// AWSDatacenterConfigStatus defines the observed state of AWSDatacenterConfig.
type AWSDatacenterConfigStatus struct { // Important: Run "make generate" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSDatacenterConfig is the Schema for the AWSDatacenterConfigs API.
type AWSDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSDatacenterConfigSpec   `json:"spec,omitempty"`
	Status AWSDatacenterConfigStatus `json:"status,omitempty"`
}

func (a *AWSDatacenterConfig) Kind() string {
	return a.TypeMeta.Kind
}

func (a *AWSDatacenterConfig) ExpectedKind() string {
	return AWSDatacenterKind
}

func (a *AWSDatacenterConfig) PauseReconcile() {
	if a.Annotations == nil {
		a.Annotations = map[string]string{}
	}
	a.Annotations[pausedAnnotation] = "true"
}

func (a *AWSDatacenterConfig) ClearPauseAnnotation() {
	if a.Annotations != nil {
		delete(a.Annotations, pausedAnnotation)
	}
}

func (a *AWSDatacenterConfig) ConvertConfigToConfigGenerateStruct() *AWSDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if a.Namespace != "" {
		namespace = a.Namespace
	}
	config := &AWSDatacenterConfigGenerate{
		TypeMeta: a.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        a.Name,
			Annotations: a.Annotations,
			Namespace:   namespace,
		},
		Spec: a.Spec,
	}

	return config
}

// +kubebuilder:object:generate=false

// Same as AWSDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig.
type AWSDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec AWSDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// AWSDatacenterConfigList contains a list of AWSDatacenterConfig.
type AWSDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSDatacenterConfig{}, &AWSDatacenterConfigList{})
}
