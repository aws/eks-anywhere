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

func (i *Image) Registry() string {
	result := strings.Split(i.URI, "/")
	if len(result) < 1 {
		return ""
	}
	return result[0]
}

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

func (i *Image) Digest() string {
	rol := strings.TrimPrefix(i.URI, i.Registry()+"/")
	result := strings.Split(rol, "@")
	if len(result) < 2 {
		return ""
	}
	return result[1]
}

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
