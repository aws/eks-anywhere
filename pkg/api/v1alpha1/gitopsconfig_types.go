package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitOps defines the configurations of GitOps Toolkit and Git repository it links to.
type GitOpsConfigSpec struct {
	Flux Flux `json:"flux,omitempty"`
}

// Flux defines the Git repository options for Flux v2.
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
	// +kubebuilder:default:="main"
	Branch string `json:"branch,omitempty"`

	// ClusterConfigPath relative to the repository root, when specified the cluster sync will be scoped to this path.
	ClusterConfigPath string `json:"clusterConfigPath,omitempty"`

	// if true, the owner is assumed to be a Git user; otherwise an org.
	Personal bool `json:"personal,omitempty"`
}

// GitOpsConfigStatus defines the observed state of GitOpsConfig.
type GitOpsConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type GitOpsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitOpsConfigSpec   `json:"spec,omitempty"`
	Status GitOpsConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=false
// Same as GitOpsConfig except stripped down for generation of yaml file while writing to github repo when flux is enabled.
type GitOpsConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec GitOpsConfigSpec `json:"spec,omitempty"`
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

// GitOpsConfigList contains a list of GitOpsConfig.
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

func (c *GitOpsConfig) ConvertToFluxConfig() *FluxConfig {
	if c == nil {
		return nil
	}
	config := &FluxConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       FluxConfigKind,
			APIVersion: c.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
		Spec: FluxConfigSpec{
			SystemNamespace:   c.Spec.Flux.Github.FluxSystemNamespace,
			Branch:            c.Spec.Flux.Github.Branch,
			ClusterConfigPath: c.Spec.Flux.Github.ClusterConfigPath,
			Github: &GithubProviderConfig{
				Owner:      c.Spec.Flux.Github.Owner,
				Repository: c.Spec.Flux.Github.Repository,
				Personal:   c.Spec.Flux.Github.Personal,
			},
		},
	}
	return config
}

func (c *GitOpsConfig) ConvertConfigToConfigGenerateStruct() *GitOpsConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &GitOpsConfigGenerate{
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

func (c *GitOpsConfig) Validate() error {
	return validateGitOpsConfig(c)
}

func (c *GitOpsConfig) SetDefaults() {
	setGitOpsConfigDefaults(c)
}

func init() {
	SchemeBuilder.Register(&GitOpsConfig{}, &GitOpsConfigList{})
}
