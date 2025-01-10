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

import "strings"

// Image represents a container image asset along with metadata such as OS,
// architecture, and registry information.
type Image struct {
	// The asset name
	Name string `json:"name,omitempty"`

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

// VersionedImage returns the full URI of the Image, including registry,
// repository, and tag or digest.
func (i Image) VersionedImage() string {
	return i.URI
}

// Image returns the repository URI of the Image, excluding the tag or digest
// if one is present.
func (i Image) Image() string {
	lastInd := strings.LastIndex(i.URI, ":")
	if lastInd == -1 {
		return i.URI
	}
	return i.URI[:lastInd]
}

// Tag returns the tag portion of the Image's URI if present, otherwise an empty string.
func (i Image) Tag() string {
	lastInd := strings.LastIndex(i.URI, ":")
	if lastInd == -1 || lastInd == len(i.URI)-1 {
		return ""
	}
	return i.URI[lastInd+1:]
}

// ChartName constructs a typical Helm chart artifact name (with ".tgz")
// from the Image's name by replacing the last colon with a hyphen.
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

// Registry returns the registry portion of the Image URI (the substring before the first slash).
func (i *Image) Registry() string {
	result := strings.Split(i.URI, "/")
	if len(result) < 1 {
		return ""
	}
	return result[0]
}

// Repository returns the repository name (between the registry and the tag/digest).
func (i *Image) Repository() string {
	rol := strings.TrimPrefix(i.URI, i.Registry()+"/")
	result := strings.Split(rol, "@")
	if len(result) < 2 {
		result = strings.Split(rol, ":")
		if len(result) < 1 {
			return ""
		}
		return result[0]
	}
	return result[0]
}

// Digest returns the SHA digest portion (after '@') of the Image URI, if present.
func (i *Image) Digest() string {
	rol := strings.TrimPrefix(i.URI, i.Registry()+"/")
	result := strings.Split(rol, "@")
	if len(result) < 2 {
		return ""
	}
	return result[1]
}

// Version returns the tag portion (after ':') of the Image URI, if present, or empty if the URI uses digests.
func (i *Image) Version() string {
	rol := strings.TrimPrefix(i.URI, i.Registry()+"/")
	result := strings.Split(rol, "@")
	if len(result) < 2 {
		result = strings.Split(rol, ":")
		if len(result) < 2 {
			return ""
		}
		return result[1]
	}
	return ""
}

// Archive represents an archive asset (e.g. tarball) along with its OS/architecture metadata,
// and checksums for file integrity.
type Archive struct {
	// The asset name
	Name string `json:"name,omitempty"`

	Description string `json:"description,omitempty"`

	// +kubebuilder:validation:Enum=linux;darwin;windows
	// Operating system of the asset
	OS string `json:"os,omitempty"`

	// +optional
	// Name of the OS like ubuntu, bottlerocket
	OSName string `json:"osName,omitempty"`

	// Architectures of the asset
	Arch []string `json:"arch,omitempty"`

	// The URI where the asset is located
	URI string `json:"uri,omitempty"`

	// The sha512 of the asset, only applies for 'file' store
	SHA512 string `json:"sha512,omitempty"`

	// The sha256 of the asset, only applies for 'file' store
	SHA256 string `json:"sha256,omitempty"`
}

// Manifest represents a reference to a manifest, typically containing
// further resource definitions or configurations.
type Manifest struct {
	// URI points to the manifest yaml file
	URI string `json:"uri,omitempty"`
}
