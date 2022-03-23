package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SnowDatacenterConfigSpec defines the desired state of SnowDatacenterConfig
type SnowDatacenterConfigSpec struct { // Important: Run "make generate" to regenerate code after modifying this file
}

// SnowDatacenterConfigStatus defines the observed state of SnowDatacenterConfig
type SnowDatacenterConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
	return SnowDatacenterKind
}

func (s *SnowDatacenterConfig) PauseReconcile() {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}
	s.Annotations[pausedAnnotation] = "true"
}

func (s *SnowDatacenterConfig) ClearPauseAnnotation() {
	if s.Annotations != nil {
		delete(s.Annotations, pausedAnnotation)
	}
}

func (s *SnowDatacenterConfig) ConvertConfigToConfigGenerateStruct() *SnowDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if s.Namespace != "" {
		namespace = s.Namespace
	}
	config := &SnowDatacenterConfigGenerate{
		TypeMeta: s.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        s.Name,
			Annotations: s.Annotations,
			Namespace:   namespace,
		},
		Spec: s.Spec,
	}

	return config
}

func (s *SnowDatacenterConfig) Marshallable() Marshallable {
	return s.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as SnowDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig
type SnowDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec SnowDatacenterConfigSpec `json:"spec,omitempty"`
}
