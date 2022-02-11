package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TinkerbellDatacenterConfigSpec defines the desired state of TinkerbellDatacenterConfig
type TinkerbellDatacenterConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file
	TinkerbellIP           string `json:"tinkerbellIP"`
	TinkerbellCertURL      string `json:"tinkerbellCertURL"`
	TinkerbellGRPCAuth     string `json:"tinkerbellGRPCAuth"`
	TinkerbellPBnJGRPCAuth string `json:"tinkerbellPBnJGRPCAuth"`
}

// TinkerbellDatacenterConfigStatus defines the observed state of TinkerbellDatacenterConfig
type TinkerbellDatacenterConfigStatus struct { // Important: Run "make generate" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TinkerbellDatacenterConfig is the Schema for the TinkerbellDatacenterConfigs API
type TinkerbellDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TinkerbellDatacenterConfigSpec   `json:"spec,omitempty"`
	Status TinkerbellDatacenterConfigStatus `json:"status,omitempty"`
}

func (t *TinkerbellDatacenterConfig) Kind() string {
	return t.TypeMeta.Kind
}

func (v *TinkerbellDatacenterConfig) ExpectedKind() string {
	return TinkerbellDatacenterKind
}

func (t *TinkerbellDatacenterConfig) PauseReconcile() {
	if t.Annotations == nil {
		t.Annotations = map[string]string{}
	}
	t.Annotations[pausedAnnotation] = "true"
}

func (t *TinkerbellDatacenterConfig) IsReconcilePaused() bool {
	if s, ok := t.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (t *TinkerbellDatacenterConfig) ClearPauseAnnotation() {
	if t.Annotations != nil {
		delete(t.Annotations, pausedAnnotation)
	}
}

func (t *TinkerbellDatacenterConfig) ConvertConfigToConfigGenerateStruct() *TinkerbellDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if t.Namespace != "" {
		namespace = t.Namespace
	}
	config := &TinkerbellDatacenterConfigGenerate{
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

func (t *TinkerbellDatacenterConfig) Marshallable() Marshallable {
	return t.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as TinkerbellDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type TinkerbellDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec TinkerbellDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// TinkerbellDatacenterConfigList contains a list of TinkerbellDatacenterConfig
type TinkerbellDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TinkerbellDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TinkerbellDatacenterConfig{}, &TinkerbellDatacenterConfigList{})
}
