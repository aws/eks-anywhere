package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// AWSIamConfig defines configuration options for AWS IAM Authenticator server

// AWSIamConfigSpec defines the desired state of AWSIamConfig
type AWSIamConfigSpec struct {
	// AWSRegion defines a region in an AWS partition
	AWSRegion string `json:"awsRegion,omitempty"`
	// BackendMode defines multiple backends for aws-iam-authenticator server
	// The server searches for mappings in order
	BackendMode []string `json:"backendMode,omitempty"`
	// ClusterID is a unique-per-cluster identifier for aws-iam-authenticator server
	ClusterID string `json:"clusterID,omitempty"`
	// MapRoles defines IAM role to a username and set of groups mapping using EKSConfigMap BackendMode
	// Each key must match AWS EKS Style ConfigMap mapRoles
	// +kubebuilder:validation:Optional
	MapRoles string `json:"mapRoles,omitempty"`
	// MapUsers defines IAM user to a username and set of groups mapping using EKSConfigMap BackendMode
	// Each key must match AWS EKS Style ConfigMap mapUsers
	// +kubebuilder:validation:Optional
	MapUsers string `json:"mapUsers,omitempty"`
	// Partition defines the AWS partition on which the IAM roles exist
	// +kubebuilder:default:=aws
	// +kubebuilder:validation:Optional
	Partition string `json:"partition,omitempty"`
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
	if e.ClusterID != n.ClusterID {
		return false
	}
	if e.MapRoles != n.MapRoles {
		return false
	}
	if e.MapUsers != n.MapUsers {
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

// AWSIamConfigStatus defines the observed state of AWSIamConfig
type AWSIamConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSIamConfig is the Schema for the awsiamconfigs API
type AWSIamConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSIamConfigSpec   `json:"spec,omitempty"`
	Status AWSIamConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=false
// Same as AWSIamConfig except stripped down for generation of yaml file while writing to github repo when flux is enabled
type AWSIamConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec AWSIamConfigSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// AWSIamConfigList contains a list of AWSIamConfig
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
	config := &AWSIamConfigGenerate{
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
	SchemeBuilder.Register(&AWSIamConfig{}, &AWSIamConfigList{})
}
