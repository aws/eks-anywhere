package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// AWSIamConfig defines configuration options for AWS IAM Authenticator server

// AWSIamConfigSpec defines the desired state of AWSIamConfig
type AWSIamConfigSpec struct {
	// BackendMode defines multiple backends for aws-iam-authenticator server in a comma separated string.
	// The server searches for mappings in order.
	BackendMode string `json:"backendMode,omitempty"`
	// ClusterID is a unique-per-cluster identifier for aws-iam-authenticator server.
	ClusterID string `json:"clusterID,omitempty"`
	// Data defines the configuration when using MountedFile BackendMode.
	// Each key must match AWS IAM Authenticator MountedFile configuration.
	// +kubebuilder:validation:Optional
	Data string `json:"data,omitempty"`
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

func (c *AWSIamConfig) Kind() string {
	return c.TypeMeta.Kind
}

func (c *AWSIamConfig) ExpectedKind() string {
	return AWSIamConfigKind
}

func init() {
	SchemeBuilder.Register(&AWSIamConfig{})
}
