package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TinkerbellMachineConfigSpec defines the desired state of TinkerbellMachineConfig.
type TinkerbellMachineConfigSpec struct {
	// HardwareSelector is a simple key-value selector for hardware.
	// Use this for straightforward single-label hardware selection.
	// Mutually exclusive with HardwareAffinity.
	// +optional
	HardwareSelector HardwareSelector `json:"hardwareSelector,omitempty"`

	// HardwareAffinity allows advanced hardware selection using required
	// and preferred affinity terms. Mutually exclusive with HardwareSelector.
	// +optional
	HardwareAffinity *HardwareAffinity `json:"hardwareAffinity,omitempty"`
	TemplateRef      Ref               `json:"templateRef,omitempty"`
	OSFamily         OSFamily          `json:"osFamily"`
	//+optional
	// OSImageURL can be used to override the default OS image path to pull from a local server.
	// OSImageURL is a URL to the OS image used during provisioning. It must include
	// the Kubernetes version(s). For example, a URL used for Kubernetes 1.27 could
	// be http://localhost:8080/ubuntu-2204-1.27.tgz
	OSImageURL          string               `json:"osImageURL"`
	Users               []UserConfiguration  `json:"users,omitempty"`
	HostOSConfiguration *HostOSConfiguration `json:"hostOSConfiguration,omitempty"`
}

// HardwareSelector models a simple key-value selector used in Tinkerbell provisioning.
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

// HardwareAffinity defines required and preferred hardware affinities.
type HardwareAffinity struct {
	// Required are the required hardware affinity terms. The terms are OR'd
	// together - hardware must match at least one term to be considered.
	// At least one required term must be specified.
	Required []HardwareAffinityTerm `json:"required"`

	// Preferred are the preferred hardware affinity terms. Hardware matching
	// these terms are preferred according to the weights provided.
	// +optional
	Preferred []WeightedHardwareAffinityTerm `json:"preferred,omitempty"`
}

// HardwareAffinityTerm defines a single hardware affinity term.
type HardwareAffinityTerm struct {
	// LabelSelector is used to select hardware by labels.
	LabelSelector metav1.LabelSelector `json:"labelSelector"`
}

// WeightedHardwareAffinityTerm is a HardwareAffinityTerm with an associated weight.
type WeightedHardwareAffinityTerm struct {
	// Weight associated with matching the corresponding term, in range 1-100.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	Weight int32 `json:"weight"`

	// HardwareAffinityTerm is the term associated with the weight.
	HardwareAffinityTerm HardwareAffinityTerm `json:"hardwareAffinityTerm"`
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

// Users returns a list of configuration for OS users.
func (c *TinkerbellMachineConfig) Users() []UserConfiguration {
	return c.Spec.Users
}

func (c *TinkerbellMachineConfig) GetNamespace() string {
	return c.Namespace
}

func (c *TinkerbellMachineConfig) GetName() string {
	return c.Name
}

// TinkerbellMachineConfigStatus defines the observed state of TinkerbellMachineConfig.
type TinkerbellMachineConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TinkerbellMachineConfig is the Schema for the tinkerbellmachineconfigs API.
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

// Validate performs light and fast Tinkerbell machine config validation.
func (c *TinkerbellMachineConfig) Validate() error {
	return validateTinkerbellMachineConfig(c)
}

// SetDefaults sets defaults for Tinkerbell machine config.
func (c *TinkerbellMachineConfig) SetDefaults() {
	setTinkerbellMachineConfigDefaults(c)
}

// +kubebuilder:object:generate=false

// Same as TinkerbellMachineConfig except stripped down for generation of yaml file during generate clusterconfig.
type TinkerbellMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec TinkerbellMachineConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// TinkerbellMachineConfigList contains a list of TinkerbellMachineConfig.
type TinkerbellMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TinkerbellMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TinkerbellMachineConfig{}, &TinkerbellMachineConfigList{})
}
