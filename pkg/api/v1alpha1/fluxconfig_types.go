// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FluxConfigSpec defines the desired state of FluxConfig.
type FluxConfigSpec struct {
	// SystemNamespace scope for this operation. Defaults to flux-system
	SystemNamespace string `json:"systemNamespace,omitempty"`

	// ClusterConfigPath relative to the repository root, when specified the cluster sync will be scoped to this path.
	ClusterConfigPath string `json:"clusterConfigPath,omitempty"`

	// Git branch. Defaults to main.
	// +kubebuilder:default:="main"
	Branch string `json:"branch,omitempty"`

	// Used to specify Github provider to host the Git repo and host the git files
	Github *GithubProviderConfig `json:"github,omitempty"`

	// Used to specify Git provider that will be used to host the git files
	Git *GitProviderConfig `json:"git,omitempty"`
}

type GithubProviderConfig struct {
	// Owner is the user or organization name of the Git provider.
	Owner string `json:"owner"`

	// Repository name.
	Repository string `json:"repository"`

	// if true, the owner is assumed to be a Git user; otherwise an org.
	Personal bool `json:"personal,omitempty"`
}

type GitProviderConfig struct {
	// Repository URL for the repository to be used with flux. Can be either an SSH or HTTPS url.
	RepositoryUrl string `json:"repositoryUrl"`

	// SSH public key algorithm for the private key specified (rsa, ecdsa, ed25519) (default ecdsa)
	SshKeyAlgorithm string `json:"sshKeyAlgorithm,omitempty"`
}

// FluxConfigStatus defines the observed state of FluxConfig.
type FluxConfigStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// FluxConfig is the Schema for the fluxconfigs API and defines the configurations of the Flux GitOps Toolkit and
// Git repository it links to.
type FluxConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FluxConfigSpec   `json:"spec,omitempty"`
	Status FluxConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:generate=false
// Same as FluxConfig except stripped down for generation of yaml file while writing to github repo when flux is enabled.
type FluxConfigGenerate struct {
	metav1.TypeMeta `json:",inline"`
	ObjectMeta      `json:"metadata,omitempty"`

	Spec FluxConfigSpec `json:"spec,omitempty"`
}

func (e *FluxConfigSpec) Equal(n *FluxConfigSpec) bool {
	if e == n {
		return true
	}
	if e == nil || n == nil {
		return false
	}
	if e.SystemNamespace != n.SystemNamespace {
		return false
	}
	if e.Branch != n.Branch {
		return false
	}
	if e.ClusterConfigPath != n.ClusterConfigPath {
		return false
	}
	return e.Git.Equal(n.Git) && e.Github.Equal(n.Github)
}

func (e *GithubProviderConfig) Equal(n *GithubProviderConfig) bool {
	if e == n {
		return true
	}
	if e == nil || n == nil {
		return false
	}
	return *e == *n
}

func (e *GitProviderConfig) Equal(n *GitProviderConfig) bool {
	if e == n {
		return true
	}
	if e == nil || n == nil {
		return false
	}
	return *e == *n
}

//+kubebuilder:object:root=true

// FluxConfigList contains a list of FluxConfig.
type FluxConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FluxConfig `json:"items"`
}

func (c *FluxConfig) Kind() string {
	return c.TypeMeta.Kind
}

func (c *FluxConfig) ExpectedKind() string {
	return FluxConfigKind
}

func (c *FluxConfig) ConvertConfigToConfigGenerateStruct() *FluxConfigGenerate {
	namespace := defaultEksaNamespace
	if c.Namespace != "" {
		namespace = c.Namespace
	}
	config := &FluxConfigGenerate{
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

func (c *FluxConfig) Validate() error {
	return validateFluxConfig(c)
}

func (c *FluxConfig) SetDefaults() {
	setFluxConfigDefaults(c)
}

func init() {
	SchemeBuilder.Register(&FluxConfig{}, &FluxConfigList{})
}
