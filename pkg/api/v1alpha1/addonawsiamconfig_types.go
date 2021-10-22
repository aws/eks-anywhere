package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// AddOnAWSIamConfig defines configuration options for AWS IAM Authenticator server

// AddOnAWSIamConfigSpec defines the desired state of AddOnAWSIamConfig
type AddOnAWSIamConfigSpec struct {
	// AWSRegion defines a region in an AWS partition
	AWSRegion string `json:"awsRegion"`
	// BackendMode defines multiple backends for aws-iam-authenticator server
	// The server searches for mappings in order
	BackendMode []string `json:"backendMode"`
	// ClusterID is a unique-per-cluster identifier for aws-iam-authenticator server
	// +kubebuilder:validation:Optional
	ClusterID string `json:"clusterID,omitempty"`
	// +kubebuilder:validation:Optional
	MapRoles []MapRoles `json:"mapRoles,omitempty"`
	// +kubebuilder:validation:Optional
	MapUsers []MapUsers `json:"mapUsers,omitempty"`
	// Partition defines the AWS partition on which the IAM roles exist
	// +kubebuilder:default:=aws
	// +kubebuilder:validation:Optional
	Partition string `json:"partition,omitempty"`
}

// MapRoles defines IAM role to a username and set of groups mapping using EKSConfigMap BackendMode
type MapRoles struct {
	RoleARN  string   `json:"roleARN"`
	Username string   `json:"username"`
	Groups   []string `json:"groups,omitempty"`
}

// MapUsers defines IAM role to a username and set of groups mapping using EKSConfigMap BackendMode
type MapUsers struct {
	UserARN  string   `json:"userARN"`
	Username string   `json:"username"`
	Groups   []string `json:"groups,omitempty"`
}

func (e *AddOnAWSIamConfigSpec) Equal(n *AddOnAWSIamConfigSpec) bool {
	if e == n {
		return true
	}
	if e == nil || n == nil {
		return false
	}
	if e.AWSRegion != n.AWSRegion {
		return false
	}
	if e.ClusterID != n.ClusterID {
		return false
	}
	if e.Partition != n.Partition {
		return false
	}
	return BackendModeSliceEqual(e.BackendMode, n.BackendMode)
}

func BackendModeSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// AddOnAWSIamConfigStatus defines the observed state of AddOnAWSIamConfig
type AddOnAWSIamConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AddOnAWSIamConfig is the Schema for the addonawsiamconfigs API
type AddOnAWSIamConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddOnAWSIamConfigSpec   `json:"spec,omitempty"`
	Status AddOnAWSIamConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=false
// Same as AddOnAWSIamConfig except stripped down for generation of yaml file while writing to github repo when flux is enabled
type AddOnAWSIamConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec AddOnAWSIamConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// AddOnAWSIamConfigList contains a list of AddOnAWSIamConfig
type AddOnAWSIamConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddOnAWSIamConfig `json:"items"`
}

func (c *AddOnAWSIamConfig) Kind() string {
	return c.TypeMeta.Kind
}

func (c *AddOnAWSIamConfig) ExpectedKind() string {
	return AddOnAWSIamConfigKind
}

func (c *AddOnAWSIamConfig) ConvertConfigToConfigGenerateStruct() *AddOnAWSIamConfigGenerate {
	config := &AddOnAWSIamConfigGenerate{
		TypeMeta: c.TypeMeta,
		ObjectMeta: ObjectMeta{
			Name:        c.Name,
			Annotations: c.Annotations,
			Namespace:   c.Namespace,
		},
		Spec: c.Spec,
	}

	return config
}

func init() {
	SchemeBuilder.Register(&AddOnAWSIamConfig{}, &AddOnAWSIamConfigList{})
}
