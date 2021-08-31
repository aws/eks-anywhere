package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitOps defines the configurations of GitOps Toolkit and Git repository it links to.
type GitOpsConfigSpec struct {
	Flux Flux `json:"flux,omitempty"`
}

// Flux defines the Git repository options for Flux v2
type Flux struct {
	// github is the name of the Git Provider to host the Git repo.
	Github Github `json:"github,omitempty"`
}

type Github struct {
	// Owner is the user or organization name of the Git provider.
	Owner string `json:"owner"`

	// Repository name.
	Repository string `json:"repository"`

	// FluxSystemNamespace scope for this operation. Defaults to flux-system.
	FluxSystemNamespace string `json:"fluxSystemNamespace,omitempty"`

	// Git branch. Defaults to main.
	Branch string `json:"branch,omitempty"`

	// ClusterConfigPath relative to the repository root, when specified the cluster sync will be scoped to this path.
	ClusterConfigPath string `json:"clusterConfigPath,omitempty"`

	// if true, the owner is assumed to be a Git user; otherwise an org.
	Personal bool `json:"personal,omitempty"`
}

// GitOpsConfigStatus defines the observed state of GitOpsConfig
type GitOpsConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type GitOpsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitOpsConfigSpec   `json:"spec,omitempty"`
	Status GitOpsConfigStatus `json:"status,omitempty"`
}

func (e *GitOpsConfigSpec) Equal(n *GitOpsConfigSpec) bool {
	if e == n {
		return true
	}
	if e == nil || n == nil {
		return false
	}
	return e.Flux == n.Flux
}

//+kubebuilder:object:root=true

// GitOpsConfigList contains a list of GitOpsConfig
type GitOpsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOpsConfig `json:"items"`
}

func (c *GitOpsConfig) Kind() string {
	return c.TypeMeta.Kind
}

func (c *GitOpsConfig) ExpectedKind() string {
	return GitOpsConfigKind
}

func init() {
	SchemeBuilder.Register(&GitOpsConfig{}, &GitOpsConfigList{})
}
