package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NutanixDatacenterConfigSpec defines the desired state of NutanixDatacenterConfig
type NutanixDatacenterConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file
	NutanixEndpoint string `json:"nutanixEndpoint"`
	NutanixPort     int    `json:"nutanixPort"`
	NutanixUser     string `json:"nutanixUser"`
	NutanixPassword string `json:"nutanixPassword"`
}

// NutanixDatacenterConfigStatus defines the observed state of NutanixDatacenterConfig
type NutanixDatacenterConfigStatus struct { // Important: Run "make generate" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NutanixDatacenterConfig is the Schema for the NutanixDatacenterConfigs API
type NutanixDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NutanixDatacenterConfigSpec   `json:"spec,omitempty"`
	Status NutanixDatacenterConfigStatus `json:"status,omitempty"`
}

func (t *NutanixDatacenterConfig) Kind() string {
	return t.TypeMeta.Kind
}

func (v *NutanixDatacenterConfig) ExpectedKind() string {
	return NutanixDatacenterKind
}

func (t *NutanixDatacenterConfig) PauseReconcile() {
	if t.Annotations == nil {
		t.Annotations = map[string]string{}
	}
	t.Annotations[pausedAnnotation] = "true"
}

func (t *NutanixDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := t.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (t *NutanixDatacenterConfig) ClearPauseAnnotation() {
	if t.Annotations != nil {
		delete(t.Annotations, pausedAnnotation)
	}
}

func (t *NutanixDatacenterConfig) ConvertConfigToConfigGenerateStruct() *NutanixDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if t.Namespace != "" {
		namespace = t.Namespace
	}
	config := &NutanixDatacenterConfigGenerate{
		TypeMeta: t.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        t.Name,
			Annotations: t.Annotations,
			Namespace:   namespace,
		},
		Spec: t.Spec,
	}

	return config
}

func (t *NutanixDatacenterConfig) Marshallable() Marshallable {
	return t.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as NutanixDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type NutanixDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec NutanixDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// NutanixDatacenterConfigList contains a list of NutanixDatacenterConfig
type NutanixDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NutanixDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NutanixDatacenterConfig{}, &NutanixDatacenterConfigList{})
}
