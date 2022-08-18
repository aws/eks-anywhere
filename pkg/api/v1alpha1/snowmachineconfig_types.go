package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

const (
	SFPPlus PhysicalNetworkConnectorType = "SFP_PLUS"
	QSFP    PhysicalNetworkConnectorType = "QSFP"

	SbeCLarge   SnowInstanceType = "sbe-c.large"
	SbeCXLarge  SnowInstanceType = "sbe-c.xlarge"
	SbeC2XLarge SnowInstanceType = "sbe-c.2xlarge"
	SbeC4XLarge SnowInstanceType = "sbe-c.4xlarge"
)

type PhysicalNetworkConnectorType string

type SnowInstanceType string

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" to regenerate code after modifying this file

// SnowMachineConfigSpec defines the desired state of SnowMachineConfigSpec
type SnowMachineConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	// The AMI ID from which to create the machine instance.
	AMIID string `json:"amiID"`

	// InstanceType is the type of instance to create.
	// Valid values: "sbe-c.large" (default), "sbe-c.xlarge", "sbe-c.2xlarge" and "sbe-c.4xlarge".
	InstanceType SnowInstanceType `json:"instanceType,omitempty"`

	// PhysicalNetworkConnector is the physical network connector type to use for creating direct network interfaces (DNI).
	// Valid values: "SFP_PLUS" (default) and "QSFP"
	PhysicalNetworkConnector PhysicalNetworkConnectorType `json:"physicalNetworkConnector,omitempty"`

	// SSHKeyName is the name of the ssh key defined in the aws snow key pairs, to attach to the instance.
	SshKeyName string `json:"sshKeyName,omitempty"`

	// Devices contains a device ip list assigned by the user to provision machines.
	Devices []string `json:"devices,omitempty"`

	// ContainersVolume provides the configuration options for the containers data storage volume.
	ContainersVolume *snowv1.Volume `json:"containersVolume,omitempty"`
}

func (s *SnowMachineConfig) SetManagedBy(clusterName string) {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}
	s.Annotations[managementAnnotation] = clusterName
}

func (s *SnowMachineConfig) OSFamily() OSFamily {
	return ""
}

// SnowMachineConfigStatus defines the observed state of SnowMachineConfig
type SnowMachineConfigStatus struct {
	// SpecValid is set to true if snow machineconfig is validated.
	SpecValid bool `json:"specValid,omitempty"`

	// FailureMessage indicates that there is a fatal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SnowMachineConfig is the Schema for the SnowMachineConfigs API
type SnowMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnowMachineConfigSpec   `json:"spec,omitempty"`
	Status SnowMachineConfigStatus `json:"status,omitempty"`
}

func (s *SnowMachineConfig) SetDefaults() {
	setSnowMachineConfigDefaults(s)
}

func (s *SnowMachineConfig) Validate() error {
	return validateSnowMachineConfig(s)
}

func (s *SnowMachineConfig) SetControlPlaneAnnotation() {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}

	s.Annotations[controlPlaneAnnotation] = "true"
}

func (s *SnowMachineConfig) SetEtcdAnnotation() {
	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}

	s.Annotations[etcdAnnotation] = "true"
}

// +kubebuilder:object:generate=false

// Same as SnowMachineConfig except stripped down for generation of yaml file during generate clusterconfig
type SnowMachineConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec SnowMachineConfigSpec `json:"spec,omitempty"`
}

func (s *SnowMachineConfig) ConvertConfigToConfigGenerateStruct() *SnowMachineConfigGenerate {
	namespace := defaultEksaNamespace
	if s.Namespace != "" {
		namespace = s.Namespace
	}
	config := &SnowMachineConfigGenerate{
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

func (s *SnowMachineConfig) Marshallable() Marshallable {
	return s.ConvertConfigToConfigGenerateStruct()
}

//+kubebuilder:object:root=true

// SnowMachineConfigList contains a list of SnowMachineConfig
type SnowMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SnowMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SnowMachineConfig{}, &SnowMachineConfigList{})
}
