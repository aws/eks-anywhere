package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DockerDatacenterConfigSpec defines the desired state of DockerDatacenterConfig.
type DockerDatacenterConfigSpec struct { // Important: Run "make generate" to regenerate code after modifying this file
	// Foo is an example field of DockerDatacenterConfig. Edit DockerDatacenter_types.go to remove/update
	// Foo string `json:"foo,omitempty"`
}

// DockerDatacenterConfigStatus defines the observed state of DockerDatacenterConfig.
type DockerDatacenterConfigStatus struct { // Important: Run "make generate" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DockerDatacenterConfig is the Schema for the DockerDatacenterConfigs API.
type DockerDatacenterConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DockerDatacenterConfigSpec   `json:"spec,omitempty"`
	Status DockerDatacenterConfigStatus `json:"status,omitempty"`
}

func (d *DockerDatacenterConfig) Kind() string {
	return d.TypeMeta.Kind
}

func (d *DockerDatacenterConfig) ExpectedKind() string {
	return DockerDatacenterKind
}

func (d *DockerDatacenterConfig) PauseReconcile() {
	if d.Annotations == nil {
		d.Annotations = map[string]string{}
	}
	d.Annotations[pausedAnnotation] = "true"
}

func (d *DockerDatacenterConfig) ClearPauseAnnotation() {
	if d.Annotations != nil {
		delete(d.Annotations, pausedAnnotation)
	}
}

func (d *DockerDatacenterConfig) ConvertConfigToConfigGenerateStruct() *DockerDatacenterConfigGenerate {
	namespace := defaultEksaNamespace
	if d.Namespace != "" {
		namespace = d.Namespace
	}
	config := &DockerDatacenterConfigGenerate{
		TypeMeta: d.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        d.Name,
			Annotations: d.Annotations,
			Namespace:   namespace,
		},
		Spec: d.Spec,
	}

	return config
}

func (d *DockerDatacenterConfig) Marshallable() Marshallable {
	return d.ConvertConfigToConfigGenerateStruct()
}

func (d *DockerDatacenterConfig) Validate() error {
	return nil
}

// +kubebuilder:object:generate=false

// Same as DockerDatacenterConfig except stripped down for generation of yaml file during generate clusterconfig.
type DockerDatacenterConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec DockerDatacenterConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// DockerDatacenterConfigList contains a list of DockerDatacenterConfig.
type DockerDatacenterConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DockerDatacenterConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DockerDatacenterConfig{}, &DockerDatacenterConfigList{})
}
