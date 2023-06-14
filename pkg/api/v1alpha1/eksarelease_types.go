package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Cluster is the Schema for the clusters API.
type EKSARelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EKSAReleaseSpec   `json:"spec,omitempty"`
	Status EKSAReleaseStatus `json:"status,omitempty"`
}

// EKSAReleaseSpec defines the desired state of EKSARelease.
type EKSAReleaseSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	Date string `json:"date"`

	// +kubebuilder:validation:Required
	// EKS-A release version
	Version string `json:"version"`

	// +kubebuilder:validation:Required
	// Git commit the component is built from, before any patches
	GitCommit string `json:"gitCommit"`

	// +kubebuilder:validation:Required
	// Manifest url to parse bundle information from for this EKS-A release
	BundleManifestUrl string `json:"bundleManifestUrl"`

	// Reference to a bundle in the cluster
	BundlesRef BundlesRef `json:"bundlesRef"`
}

// EKSAReleaseStatus defines the observed state of EKSARelease.
type EKSAReleaseStatus struct{}

// +kubebuilder:object:root=true
// EKSAReleaseList contains a list of EKSARelease.
type EKSAReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EKSARelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EKSARelease{}, &EKSAReleaseList{})
}
