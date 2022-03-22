package v1alpha1

import "strings"

type Image struct {
	// +kubebuilder:validation:Required
	// The asset name
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Enum=linux;darwin;windows
	// Operating system of the asset
	OS string `json:"os,omitempty"`

	// +optional
	// Name of the OS like ubuntu, bottlerocket
	OSName string `json:"osName,omitempty"`

	// Architectures of the asset
	Arch []string `json:"arch,omitempty"`

	// The image repository, name, and tag
	URI string `json:"uri,omitempty"`

	// The SHA256 digest of the image manifest
	ImageDigest string `json:"imageDigest,omitempty"`
}

func (i Image) VersionedImage() string {
	return i.URI
}

func (i Image) Image() string {
	lastInd := strings.LastIndex(i.URI, ":")
	if lastInd == -1 {
		return i.URI
	}
	return i.URI[:lastInd]
}

func (i Image) Tag() string {
	lastInd := strings.LastIndex(i.URI, ":")
	if lastInd == -1 || lastInd == len(i.URI)-1 {
		return ""
	}
	return i.URI[lastInd+1:]
}

func (i Image) ChartName() string {
	lastInd := strings.LastIndex(i.Image(), "/")
	if lastInd == -1 {
		return i.URI
	}
	chart := i.URI[lastInd+1:]
	chart = strings.Replace(chart, ":", "-", 1)
	chart += ".tgz"
	return chart
}

type Archive struct {
	// +kubebuilder:validation:Required
	// The asset name
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Enum=linux;darwin;windows
	// Operating system of the asset
	OS string `json:"os,omitempty"`

	// +optional
	// Name of the OS like ubuntu, bottlerocket
	OSName string `json:"osName,omitempty"`

	// Architectures of the asset
	Arch []string `json:"arch,omitempty"`

	// +kubebuilder:validation:Required
	// The URI where the asset is located
	URI string `json:"uri,omitempty"`
	// +kubebuilder:validation:Required
	// The sha512 of the asset, only applies for 'file' store
	SHA512 string `json:"sha512,omitempty"`
	// +kubebuilder:validation:Required
	// The sha256 of the asset, only applies for 'file' store
	SHA256 string `json:"sha256,omitempty"`
}

type Manifest struct {
	// +kubebuilder:validation:Required
	// URI points to the manifest yaml file
	URI string `json:"uri,omitempty"`
}
