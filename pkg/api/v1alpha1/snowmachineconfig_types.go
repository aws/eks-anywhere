package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SnowMachineConfigSpec defines the desired state of SnowMachineConfigSpec
type SnowMachineConfigSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	// The AMI ID from which to create the machine instance.
	AMIID string `json:"amiID"`

	// InstanceType is the type of instance to create. Example: m4.xlarge
	InstanceType string `json:"instanceType,omitempty"`

	// PhysicalNetworkConnector is the physical network connector type to use for creating direct network interfaces (DNI).
	// Valid values: "SFP_PLUS" (default) and "QSFP"
	PhysicalNetworkConnector string `json:"physicalNetworkConnector,omitempty"`

	// SSHKeyName is the name of the ssh key defined in the aws snow key pairs, to attach to the instance.
	SshKeyName string `json:"sshKeyName,omitempty"`
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
type SnowMachineConfigStatus struct{}

// SnowMachineConfig is the Schema for the SnowMachineConfigs API
type SnowMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnowMachineConfigSpec   `json:"spec,omitempty"`
	Status SnowMachineConfigStatus `json:"status,omitempty"`
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
