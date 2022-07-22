package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TinkerbellMachineConfigSpec defines the desired state of TinkerbellMachineConfig
type TinkerbellMachineConfigSpec struct {
	HardwareSelector HardwareSelector    `json:"hardwareSelector"`
	TemplateRef      Ref                 `json:"templateRef,omitempty"`
	OSFamily         OSFamily            `json:"osFamily"`
	Users            []UserConfiguration `json:"users,omitempty"`
}

// HardwareSelector models a simple key-value selector used in Tinkerbell providioning.
type HardwareSelector map[string]string

// IsEmpty returns true if s has no key-value pairs.
func (s HardwareSelector) IsEmpty() bool {
	return len(s) == 0
}

func (s HardwareSelector) ToString() (string, error) {
	encoded, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func (c *TinkerbellMachineConfig) PauseReconcile() {
	c.Annotations[pausedAnnotation] = "true"
}

func (c *TinkerbellMachineConfig) IsReconcilePaused() bool {
	if s, ok := c.Annotations[pausedAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *TinkerbellMachineConfig) SetControlPlane() {
	c.Annotations[controlPlaneAnnotation] = "true"
}

func (c *TinkerbellMachineConfig) IsControlPlane() bool {
	if s, ok := c.Annotations[controlPlaneAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *TinkerbellMachineConfig) SetEtcd() {
	c.Annotations[etcdAnnotation] = "true"
}

func (c *TinkerbellMachineConfig) IsEtcd() bool {
	if s, ok := c.Annotations[etcdAnnotation]; ok {
		return s == "true"
	}
	return false
}

func (c *TinkerbellMachineConfig) SetManagedBy(clusterName string) {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[managementAnnotation] = clusterName
}

func (c *TinkerbellMachineConfig) IsManaged() bool {
	if s, ok := c.Annotations[managementAnnotation]; ok {
		return s != ""
	}
	return false
}

func (c *TinkerbellMachineConfig) OSFamily() OSFamily {
	return c.Spec.OSFamily
}

func (c *TinkerbellMachineConfig) GetNamespace() string {
	return c.Namespace
}

func (c *TinkerbellMachineConfig) GetName() string {
	return c.Name
}

// TinkerbellMachineConfigStatus defines the observed state of TinkerbellMachineConfig
type TinkerbellMachineConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TinkerbellMachineConfig is the Schema for the tinkerbellmachineconfigs API
type TinkerbellMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TinkerbellMachineConfigSpec   `json:"spec,omitempty"`
	Status TinkerbellMachineConfigStatus `json:"status,omitempty"`
}

func (c *TinkerbellMachineConfig) ConvertConfigToConfigGenerateStruct() *TinkerbellMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &TinkerbellMachineConfigGenerate{
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

func (c *TinkerbellMachineConfig) Marshallable() Marshallable {
	return c.ConvertConfigToConfigGenerateStruct()
}

// +kubebuilder:object:generate=false

// Same as TinkerbellMachineConfig except stripped down for generation of yaml file during generate clusterconfig
type TinkerbellMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec TinkerbellMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// TinkerbellMachineConfigList contains a list of TinkerbellMachineConfig
type TinkerbellMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TinkerbellMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TinkerbellMachineConfig{}, &TinkerbellMachineConfigList{})
}
