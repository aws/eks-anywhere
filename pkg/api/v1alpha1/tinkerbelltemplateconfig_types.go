package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" to regenerate code after modifying this file

// TinkerbellTemplateConfigSpec defines the desired state of TinkerbellTemplateConfig.
type TinkerbellTemplateConfigSpec struct {
	// Template defines a Tinkerbell workflow template with specific tasks and actions.
	Template tinkerbell.Workflow `json:"template"`
}

// TinkerbellTemplateConfigStatus defines the observed state of TinkerbellTemplateConfig.
type TinkerbellTemplateConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TinkerbellTemplateConfig is the Schema for the TinkerbellTemplateConfigs API.
type TinkerbellTemplateConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TinkerbellTemplateConfigSpec   `json:"spec,omitempty"`
	Status TinkerbellTemplateConfigStatus `json:"status,omitempty"`
}

func (t *TinkerbellTemplateConfig) Kind() string {
	return t.TypeMeta.Kind
}

func (t *TinkerbellTemplateConfig) ExpectedKind() string {
	return TinkerbellTemplateConfigKind
}

func (t *TinkerbellTemplateConfig) ToTemplateString() (string, error) {
	b, err := yaml.Marshal(&t.Spec.Template)
	if err != nil {
		return "", fmt.Errorf("failed to convert TinkerbellTemplateConfig.Spec.Template to string: %v", err)
	}
	return string(b), nil
}

func (c *TinkerbellTemplateConfig) ConvertConfigToConfigGenerateStruct() *TinkerbellTemplateConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &TinkerbellTemplateConfigGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   namespace,
		},
		Spec: c.Spec,
	}

	return config
}

// +kubebuilder:object:generate=false

// Same as TinkerbellTemplateConfig except stripped down for generation of yaml file during generate clusterconfig.
type TinkerbellTemplateConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec TinkerbellTemplateConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// TinkerbellTemplateConfigList contains a list of TinkerbellTemplateConfig.
type TinkerbellTemplateConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TinkerbellTemplateConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TinkerbellTemplateConfig{}, &TinkerbellTemplateConfigList{})
}
