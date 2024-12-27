/*
Copyright 2022 The Tinkerbell Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ClusterFinalizer allows ReconcileTinkerbellCluster to clean up Tinkerbell resources before
	// removing it from the apiserver.
	ClusterFinalizer = "tinkerbellcluster.infrastructure.cluster.x-k8s.io"
)

// TinkerbellClusterSpec defines the desired state of TinkerbellCluster.
type TinkerbellClusterSpec struct {
	// ControlPlaneEndpoint is a required field by ClusterAPI v1beta1.
	//
	// See https://cluster-api.sigs.k8s.io/developer/architecture/controllers/cluster.html
	// for more details.
	//
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`

	// ImageLookupFormat is the URL naming format to use for machine images when
	// a machine does not specify. When set, this will be used for all cluster machines
	// unless a machine specifies a different ImageLookupFormat. Supports substitutions
	// for {{.BaseRegistry}}, {{.OSDistro}}, {{.OSVersion}} and {{.KubernetesVersion}} with
	// the basse URL, OS distribution, OS version, and kubernetes version, respectively.
	// BaseRegistry will be the value in ImageLookupBaseRegistry or ghcr.io/tinkerbell/cluster-api-provider-tinkerbell
	// (the default), OSDistro will be the value in ImageLookupOSDistro or ubuntu (the default),
	// OSVersion will be the value in ImageLookupOSVersion or default based on the OSDistro
	// (if known), and the kubernetes version as defined by the packages produced by
	// kubernetes/release: v1.13.0, v1.12.5-mybuild.1, or v1.17.3. For example, the default
	// image format of {{.BaseRegistry}}/{{.OSDistro}}-{{.OSVersion}}:{{.KubernetesVersion}}.gz will
	// attempt to pull the image from that location. See also: https://golang.org/pkg/text/template/
	// +optional
	ImageLookupFormat string `json:"imageLookupFormat,omitempty"`

	// ImageLookupBaseRegistry is the base Registry URL that is used for pulling images,
	// if not set, the default will be to use ghcr.io/tinkerbell/cluster-api-provider-tinkerbell.
	// +optional
	// +kubebuilder:default=ghcr.io/tinkerbell/cluster-api-provider-tinkerbell
	ImageLookupBaseRegistry string `json:"imageLookupBaseRegistry,omitempty"`

	// ImageLookupOSDistro is the name of the OS distro to use when fetching machine images,
	// if not set it will default to ubuntu.
	// +optional
	// +kubebuilder:default=ubuntu
	ImageLookupOSDistro string `json:"imageLookupOSDistro,omitempty"`

	// ImageLookupOSVersion is the version of the OS distribution to use when fetching machine
	// images. If not set it will default based on ImageLookupOSDistro.
	// +optional
	ImageLookupOSVersion string `json:"imageLookupOSVersion,omitempty"`
}

// TinkerbellClusterStatus defines the observed state of TinkerbellCluster.
type TinkerbellClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file.

	// Ready denotes that the cluster (infrastructure) is ready.
	// +optional
	Ready bool `json:"ready"`
}

// +kubebuilder:subresource:status
// +kubebuilder:resource:path=tinkerbellclusters,scope=Namespaced,categories=cluster-api
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this TinkerbellCluster belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="TinkerbellCluster ready status"

// TinkerbellCluster is the Schema for the tinkerbellclusters API.
type TinkerbellCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TinkerbellClusterSpec   `json:"spec,omitempty"`
	Status TinkerbellClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TinkerbellClusterList contains a list of TinkerbellCluster.
type TinkerbellClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TinkerbellCluster `json:"items"`
}

//nolint:gochecknoinits
func init() {
	SchemeBuilder.Register(&TinkerbellCluster{}, &TinkerbellClusterList{})
}
