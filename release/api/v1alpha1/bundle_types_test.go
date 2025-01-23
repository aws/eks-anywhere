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

//nolint:revive
import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestBundlesDefaultEksAToolsImage(t *testing.T) {
	g := NewWithT(t)
	bundles := &v1alpha1.Bundles{
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					Eksa: v1alpha1.EksaBundle{
						CliTools: v1alpha1.Image{
							URI: "tools:v1.0.0",
						},
					},
				},
			},
		},
	}
	g.Expect(bundles.DefaultEksAToolsImage()).To(Equal(v1alpha1.Image{URI: "tools:v1.0.0"}))
}
