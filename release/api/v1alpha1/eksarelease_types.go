package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EKSARelease is the mapping between release semver of EKS-A and a Bundles resource on the cluster.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type EKSARelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec EKSAReleaseSpec `json:"spec,omitempty"`
}

// EKSAReleaseSpec defines the desired state of EKSARelease.
type EKSAReleaseSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// Date of EKS-A Release
	ReleaseDate string `json:"releaseDate"`

	// +kubebuilder:validation:Required
	// EKS-A release semantic version
	Version string `json:"version"`

	// +kubebuilder:validation:Required
	// Git commit the component is built from, before any patches
	GitCommit string `json:"gitCommit"`

	// +kubebuilder:validation:Required
	// Manifest url to parse bundle information from for this EKS-A release
	BundleManifestURL string `json:"bundleManifestUrl"`

	// Reference to a Bundles resource in the cluster
	BundlesRef BundlesRef `json:"bundlesRef"`
}

// EKSAReleaseStatus defines the observed state of EKSARelease.
type EKSAReleaseStatus struct{}

// BundlesRef refers to a Bundles resource in the cluster.
type BundlesRef struct {
	// APIVersion refers to the Bundles APIVersion
	APIVersion string `json:"apiVersion"`
	// Name refers to the name of the Bundles object in the cluster
	Name string `json:"name"`
	// Namespace refers to the Bundles's namespace
	Namespace string `json:"namespace"`
}

// EKSAReleaseList is a list of EKSARelease resources.
// +kubebuilder:object:root=true
type EKSAReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []EKSARelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EKSARelease{}, &EKSAReleaseList{})
}
