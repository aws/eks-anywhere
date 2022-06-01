package v1alpha1

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NutanixDatacenterConfigSpec defines the desired state of NutanixDatacenterConfig
type NutanixDatacenterConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file
	// NutanixEndpoint is the Endpoint of Nutanix Prism Central
	// +kubebuilder:validation:Required
	NutanixEndpoint string `json:"nutanixEndpoint"`
	// NutanixPort is the Port of Nutanix Prism Central
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=9440
	NutanixPort int `json:"nutanixPort"`
	// NutanixUser is the User name for Nutanix Prism Central
	// +kubebuilder:validation:Required
	NutanixUser string `json:"nutanixUser"`
	// NutanixPassword is the Password for Nutanix Prism Central
	// +kubebuilder:validation:Required
	NutanixPassword string `json:"nutanixPassword"`
	// NutanixInsecure is the protocol type to connect to Nutanix Prism Central
	// +kubebuilder:validation:Required
	NutanixInsecure bool `json:"nutanixInsecure"`
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

func (t *NutanixDatacenterConfig) Validate() error {
	if len(t.Spec.NutanixEndpoint) <= 0 {
		return errors.New("NutanixDatacenterConfig nutanixEndpoint is not set or is empty")
	}

	if t.Spec.NutanixPort == 0 {
		return errors.New("NutanixDatacenterConfig nutanixPort is not set or is empty")
	}

	if len(t.Spec.NutanixUser) <= 0 {
		return errors.New("NutanixDatacenterConfig nutanixUser is not set or is empty")
	}

	if len(t.Spec.NutanixPassword) <= 0 {
		return errors.New("NutanixDatacenterConfig nutanixPassword is not set or is empty")
	}
	return nil
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

//var _ providers.DatacenterConfig = = (*NutanixDatacenterConfig)(nil)
func init() {
	SchemeBuilder.Register(&NutanixDatacenterConfig{}, &NutanixDatacenterConfigList{})
}
