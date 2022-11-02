package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// AWSIamConfig defines configuration options for AWS IAM Authenticator server

// AWSIamConfigSpec defines the desired state of AWSIamConfig.
type AWSIamConfigSpec struct {
	// AWSRegion defines a region in an AWS partition
	AWSRegion string `json:"awsRegion"`
	// BackendMode defines multiple backends for aws-iam-authenticator server
	// The server searches for mappings in order
	BackendMode []string `json:"backendMode"`
	// +kubebuilder:validation:Optional
	MapRoles []MapRoles `json:"mapRoles,omitempty"`
	// +kubebuilder:validation:Optional
	MapUsers []MapUsers `json:"mapUsers,omitempty"`
	// Partition defines the AWS partition on which the IAM roles exist
	// +kubebuilder:default:=aws
	// +kubebuilder:validation:Optional
	Partition string `json:"partition,omitempty"`
}

// MapRoles defines IAM role to a username and set of groups mapping using EKSConfigMap BackendMode.
type MapRoles struct {
	RoleARN  string   `yaml:"rolearn" json:"roleARN"`
	Username string   `json:"username"`
	Groups   []string `json:"groups,omitempty"`
}

// MapUsers defines IAM role to a username and set of groups mapping using EKSConfigMap BackendMode.
type MapUsers struct {
	UserARN  string   `yaml:"userarn" json:"userARN"`
	Username string   `json:"username"`
	Groups   []string `json:"groups,omitempty"`
}

func (e *AWSIamConfigSpec) Equal(n *AWSIamConfigSpec) bool {
	if e == n {
		return true
	}
	if e == nil || n == nil {
		return false
	}
	if e.AWSRegion != n.AWSRegion {
		return false
	}
	if e.Partition != n.Partition {
		return false
	}
	return SliceEqual(e.BackendMode, n.BackendMode)
}

// AWSIamConfigStatus defines the observed state of AWSIamConfig.
type AWSIamConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSIamConfig is the Schema for the awsiamconfigs API.
type AWSIamConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSIamConfigSpec   `json:"spec,omitempty"`
	Status AWSIamConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=false
// Same as AWSIamConfig except stripped down for generation of yaml file while writing to github repo when flux is enabled.
type AWSIamConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec AWSIamConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// AWSIamConfigList contains a list of AWSIamConfig.
type AWSIamConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSIamConfig `json:"items"`
}

func (c *AWSIamConfig) Kind() string {
	return c.TypeMeta.Kind
}

func (c *AWSIamConfig) ExpectedKind() string {
	return AWSIamConfigKind
}

func (c *AWSIamConfig) ConvertConfigToConfigGenerateStruct() *AWSIamConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &AWSIamConfigGenerate{
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

func (c *AWSIamConfig) Validate() error {
	return validateAWSIamConfig(c)
}

func (c *AWSIamConfig) SetDefaults() {
	setDefaultAWSIamPartition(c)
}

func init() {
	SchemeBuilder.Register(&AWSIamConfig{}, &AWSIamConfigList{})
}
