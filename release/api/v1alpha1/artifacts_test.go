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

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestVersionsBundleSnowImages(t *testing.T) {
	tests := []struct {
		name           string
		versionsBundle *v1alpha1.VersionsBundle
		want           []v1alpha1.Image
	}{
		{
			name:           "no images",
			versionsBundle: &v1alpha1.VersionsBundle{},
			want:           []v1alpha1.Image{},
		},
		{
			name: "kubevip images",
			versionsBundle: &v1alpha1.VersionsBundle{
				Snow: v1alpha1.SnowBundle{
					KubeVip: v1alpha1.Image{
						Name: "kubevip",
						URI:  "uri",
					},
				},
			},
			want: []v1alpha1.Image{
				{
					Name: "kubevip",
					URI:  "uri",
				},
			},
		},
		{
			name: "manager images",
			versionsBundle: &v1alpha1.VersionsBundle{
				Snow: v1alpha1.SnowBundle{
					Manager: v1alpha1.Image{
						Name: "manage",
						URI:  "uri",
					},
				},
			},
			want: []v1alpha1.Image{
				{
					Name: "manage",
					URI:  "uri",
				},
			},
		},
		{
			name: "bootstrap-snow images",
			versionsBundle: &v1alpha1.VersionsBundle{
				Snow: v1alpha1.SnowBundle{
					Manager: v1alpha1.Image{
						Name: "bootstrap-snow",
						URI:  "uri",
					},
				},
			},
			want: []v1alpha1.Image{
				{
					Name: "bootstrap-snow",
					URI:  "uri",
				},
			},
		},
		{
			name: "all images",
			versionsBundle: &v1alpha1.VersionsBundle{
				Snow: v1alpha1.SnowBundle{
					KubeVip: v1alpha1.Image{
						Name: "kubevip",
						URI:  "uri",
					},
					Manager: v1alpha1.Image{
						Name: "manage",
						URI:  "uri",
					},
					BottlerocketBootstrapSnow: v1alpha1.Image{
						Name: "bootstrap-snow",
						URI:  "uri",
					},
				},
			},
			want: []v1alpha1.Image{
				{
					Name: "kubevip",
					URI:  "uri",
				},
				{
					Name: "manage",
					URI:  "uri",
				},
				{
					Name: "bootstrap-snow",
					URI:  "uri",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.versionsBundle.SnowImages()).To(Equal(tt.want))
		})
	}
}
