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

package v1alpha1_test

import (
	"testing"

	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestImageVersionedImage(t *testing.T) {
	tests := []struct {
		testName string
		URI      string
		want     string
	}{
		{
			testName: "full uri",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
		},
		{
			testName: "full uri with port",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			i := v1alpha1.Image{
				URI: tt.URI,
			}
			if got := i.VersionedImage(); got != tt.want {
				t.Errorf("Image.VersionedImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageImage(t *testing.T) {
	tests := []struct {
		testName string
		URI      string
		want     string
	}{
		{
			testName: "full uri",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node",
		},
		{
			testName: "full uri with port",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node",
		},
		{
			testName: "no tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node",
			want:     "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			i := v1alpha1.Image{
				URI: tt.URI,
			}
			if got := i.Image(); got != tt.want {
				t.Errorf("Image.Image() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageTag(t *testing.T) {
	tests := []struct {
		testName string
		URI      string
		want     string
	}{
		{
			testName: "full uri",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
		},
		{
			testName: "full uri with port",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
		},
		{
			testName: "no tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node",
			want:     "",
		},
		{
			testName: "empty tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			i := v1alpha1.Image{
				URI: tt.URI,
			}
			if got := i.Tag(); got != tt.want {
				t.Errorf("Image.Tag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImage_Registry(t *testing.T) {
	tests := []struct {
		testName string
		URI      string
		want     string
	}{
		{
			testName: "full uri",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "public.ecr.aws",
		},
		{
			testName: "full uri with port",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "public.ecr.aws:8484",
		},
		{
			testName: "no slash",
			URI:      "public.ecr.aws",
			want:     "public.ecr.aws",
		},
		{
			testName: "nothing",
			URI:      "",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			i := v1alpha1.Image{
				URI: tt.URI,
			}
			if got := i.Registry(); got != tt.want {
				t.Errorf("Image.Registry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImage_Repository(t *testing.T) {
	tests := []struct {
		testName string
		URI      string
		want     string
	}{
		{
			testName: "tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "l0g8r8j6/kubernetes-sigs/kind/node",
		},
		{
			testName: "port and tag",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "l0g8r8j6/kubernetes-sigs/kind/node",
		},
		{
			testName: "port and sha256",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node@sha256:6165d26ef648100226c1944c6b1c83e875a4bf81bba91054a00c5121cfeff363",
			want:     "l0g8r8j6/kubernetes-sigs/kind/node",
		},
		{
			testName: "port no tag",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node",
			want:     "l0g8r8j6/kubernetes-sigs/kind/node",
		},
		{
			testName: "no tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node",
			want:     "l0g8r8j6/kubernetes-sigs/kind/node",
		},
		{
			testName: "no nothing",
			URI:      "",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			i := v1alpha1.Image{
				URI: tt.URI,
			}
			if got := i.Repository(); got != tt.want {
				t.Errorf("Image.Repository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImage_Version(t *testing.T) {
	tests := []struct {
		testName string
		URI      string
		want     string
	}{
		{
			testName: "tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
		},
		{
			testName: "port and tag",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
		},
		{
			testName: "port and sha256",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node@sha256:6165d26ef648100226c1944c6b1c83e875a4bf81bba91054a00c5121cfeff363",
			want:     "",
		},
		{
			testName: "port no tag",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node",
			want:     "",
		},
		{
			testName: "no tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node",
			want:     "",
		},
		{
			testName: "no nothing",
			URI:      "",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			i := v1alpha1.Image{
				URI: tt.URI,
			}
			if got := i.Version(); got != tt.want {
				t.Errorf("Image.Version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImage_Digest(t *testing.T) {
	tests := []struct {
		testName string
		URI      string
		want     string
	}{
		{
			testName: "tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node:v1.20.4-eks-d-1-20-1-eks-a-0.0.1.build.38",
			want:     "",
		},
		{
			testName: "port and sha256",
			URI:      "public.ecr.aws:8484/l0g8r8j6/kubernetes-sigs/kind/node@sha256:6165d26ef648100226c1944c6b1c83e875a4bf81bba91054a00c5121cfeff363",
			want:     "sha256:6165d26ef648100226c1944c6b1c83e875a4bf81bba91054a00c5121cfeff363",
		},
		{
			testName: "no tag",
			URI:      "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/node",
			want:     "",
		},
		{
			testName: "no nothing",
			URI:      "",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			i := v1alpha1.Image{
				URI: tt.URI,
			}
			if got := i.Digest(); got != tt.want {
				t.Errorf("Image.Digest() = %v, want %v", got, tt.want)
			}
		})
	}
}
