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

// ReleaseSpec defines the desired state of Release.
type ReleaseSpec struct {
	// +kubebuilder:validation:Required
	// EKS-A Latest Release version following semver
	LatestVersion string `json:"latestVersion"`

	// +kubebuilder:validation:Required
	// List of all eks-a releases
	Releases []EksARelease `json:"releases"`
}

// ReleaseStatus defines the observed state of Release.
type ReleaseStatus struct{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Release is the Schema for the releases API.
type Release struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseSpec   `json:"spec"`
	Status ReleaseStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReleaseList contains a list of Release.
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Release `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Release{}, &ReleaseList{})
}

// EksARelease defines each release of EKS-Anywhere.
type EksARelease struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	Date string `json:"date"`

	// +kubebuilder:validation:Required
	// EKS-A release version
	Version string `json:"version"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// Monotonically increasing release number
	Number int `json:"number"`

	// +kubebuilder:validation:Required
	// Git commit the component is built from, before any patches
	GitCommit string `json:"gitCommit"`

	// Git tag the component is built from, before any patches
	GitTag string `json:"gitTag"`

	// +kubebuilder:validation:Required
	// Manifest url to parse bundle information from for this EKS-A release
	BundleManifestUrl string `json:"bundleManifestUrl"`

	// +kubebuilder:validation:Required
	// EKS Anywhere binary bundle
	EksABinary BinaryBundle `json:"eksABinary"`

	// +kubebuilder:validation:Required
	// EKS Anywhere CLI bundle
	EksACLI PlatformBundle `json:"eksACLI"`
}

type BinaryBundle struct {
	// +kubebuilder:validation:Required
	// EKS Anywhere Linux binary
	LinuxBinary Archive `json:"linux"`

	// +kubebuilder:validation:Required
	// EKS Anywhere Darwin binary
	DarwinBinary Archive `json:"darwin"`
}

type PlatformBundle struct {
	// +kubebuilder:validation:Required
	// EKS Anywhere Linux binary
	LinuxBinary ArchitectureBundle `json:"linux"`

	// +kubebuilder:validation:Required
	// EKS Anywhere Darwin binary
	DarwinBinary ArchitectureBundle `json:"darwin"`
}

type ArchitectureBundle struct {
	Amd64 Archive `json:"amd64"`
	Arm64 Archive `json:"arm64"`
}
