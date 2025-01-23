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

package signature

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
)

func TestGetBundleSignature(t *testing.T) {
	testCases := []struct {
		testName        string
		bundle          *anywherev1alpha1.Bundles
		key             string
		expectErrSubstr string
	}{
		{
			testName:        "Nil bundle",
			bundle:          nil,
			key:             constants.KmsKey,
			expectErrSubstr: "computing digest:",
		},
		{
			testName: "Excluding fields from bundle with minimal valid data",
			bundle: &anywherev1alpha1.Bundles{
				Spec: anywherev1alpha1.BundlesSpec{
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			key:             constants.KmsKey,
			expectErrSubstr: "",
		},
		{
			testName: "Excluding fields from a fully populated Bundles object",
			bundle: &anywherev1alpha1.Bundles{
				Spec: anywherev1alpha1.BundlesSpec{
					Number:        10,
					CliMinVersion: "v1.0.0",
					CliMaxVersion: "v2.0.0",
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion:          "1.28",
							EndOfStandardSupport: "2024-10-10",
							EksD: anywherev1alpha1.EksDRelease{
								Name:           "eks-d-1-25",
								ReleaseChannel: "1-25",
								KubeVersion:    "1.25",
								EksDReleaseUrl: "https://example.com/release.yaml",
							},
							CertManager: anywherev1alpha1.CertManagerBundle{
								Version: "v1.11.0",
								Acmesolver: anywherev1alpha1.Image{
									URI: "public.ecr.aws/acmesolver:latest",
								},
								Cainjector: anywherev1alpha1.Image{
									URI: "public.ecr.aws/cainjector:latest",
								},
								Controller: anywherev1alpha1.Image{
									URI: "public.ecr.aws/cert-manager-controller:latest",
								},
								Startupapicheck: anywherev1alpha1.Image{
									URI: "public.ecr.aws/startupapicheck:latest",
								},
								Webhook: anywherev1alpha1.Image{
									URI: "public.ecr.aws/webhook:latest",
								},
								Manifest: anywherev1alpha1.Manifest{
									URI: "https://example.com/cert-manager.yaml",
								},
							},
							Eksa: anywherev1alpha1.EksaBundle{
								Version: "v0.0.1-dev",
								CliTools: anywherev1alpha1.Image{
									URI: "public.ecr.aws/eks-anywhere-cli-tools:latest",
								},
							},
						},
						{
							KubeVersion: "1.26",
						},
					},
				},
			},
			key:             constants.KmsKey,
			expectErrSubstr: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)

			ctx := context.Background()
			sig, err := GetBundleSignature(ctx, tt.bundle, tt.key)

			if tt.expectErrSubstr == "" {
				// Expecting no particular error substring -> test for success
				g.Expect(err).NotTo(HaveOccurred(),
					"Expected no error but got error: %v", err)
				g.Expect(sig).NotTo(BeEmpty(),
					"Expected signature string to be non-empty on success")
			} else {
				// Expecting an error substring -> test for error presence
				g.Expect(err).To(HaveOccurred(),
					"Expected an error but got none")
				g.Expect(err.Error()).To(ContainSubstring(tt.expectErrSubstr),
					"Error message should contain substring %q, got: %v", tt.expectErrSubstr, err)
				g.Expect(sig).To(BeEmpty(),
					"Expected signature to be empty when error occurs")
			}
		})
	}
}

func TestGetDigest(t *testing.T) {
	testCases := []struct {
		testName        string
		bundle          *anywherev1alpha1.Bundles
		expectErrSubstr string
	}{
		{
			testName: "Simple valid bundle",
			bundle: &anywherev1alpha1.Bundles{
				Spec: anywherev1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			expectErrSubstr: "",
		},
		{
			testName: "Another valid bundle with more fields",
			bundle: &anywherev1alpha1.Bundles{
				Spec: anywherev1alpha1.BundlesSpec{
					Number:        10,
					CliMinVersion: "v0.0.1",
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.28",
						},
						{
							KubeVersion: "1.29",
						},
					},
				},
			},
			expectErrSubstr: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)

			digest, filtered, err := getDigest(tt.bundle)
			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(HaveOccurred(), "Expected success but got error")
				g.Expect(digest).NotTo(BeZero(),
					"Expected digest to be non-zero array")
				g.Expect(filtered).NotTo(BeEmpty(),
					"Expected filtered bytes to be non-empty")
			} else {
				g.Expect(err).To(HaveOccurred(),
					"Expected error but got none")
				g.Expect(err.Error()).To(ContainSubstring(tt.expectErrSubstr),
					"Error message should contain substring %q, got: %v",
					tt.expectErrSubstr, err)
				g.Expect(digest).To(BeZero())
				g.Expect(filtered).To(BeNil())
			}
		})
	}
}

func TestFilterExcludes(t *testing.T) {
	testCases := []struct {
		testName        string
		jsonPayload     string
		expectErrSubstr string
		expectExclude   []string // substrings we expect to NOT be present
		expectInclude   []string // substrings we expect to be present
	}{
		{
			testName: "Valid JSON with known excludes",
			jsonPayload: `{
				"metadata": {
					"creationTimestamp": "2021-09-01T00:00:00Z",
					"annotations": { "key": "value" }
				},
				"status": {
					"someStatus": "info"
				},
				"spec": {
					"versionsBundles": [{
						"kubeVersion": "1.28",
						"endOfExtendedSupport": "2024-12-31",
						"eksD": {
							"channel": "1-28",
							"components": "https://distro.eks.amazonaws.com/crds/releases.distro.eks.amazonaws.com-v1alpha1.yaml",
							"gitCommit": "3c3ff5d3aaa7417b906549756da44f60af5df03d",
							"kubeVersion": "v1.28.15",
							"manifestUrl": "https://distro.eks.amazonaws.com/kubernetes-1-28/kubernetes-1-28-eks-37.yaml",
							"name": "kubernetes-1-28-eks-37"
						},
						"eksa": "someValue"
					}],
					"otherField": "otherValue"
				}
			}`,
			expectErrSubstr: "",
			expectExclude: []string{
				"creationTimestamp",
				"annotations",
				"status",
				"eksa",
			},
			expectInclude: []string{
				"kubeVersion",
				"endOfExtendedSupport",
				"eksD",
			},
		},
		{
			testName:        "Invalid JSON payload",
			jsonPayload:     `{"unclosed": [`,
			expectErrSubstr: "unmarshalling JSON:",
		},
		{
			testName: "Excludes with minimal JSON",
			jsonPayload: `{
				"metadata": {"creationTimestamp": "2021-09-01T00:00:00Z"},
				"spec": {
					"versionsBundles": [{
						"kubeVersion": "1.31"
					}]
				}
			}`,
			expectErrSubstr: "",
			expectExclude:   []string{"creationTimestamp"},
			expectInclude:   []string{"spec"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)

			filtered, err := filterExcludes([]byte(tt.jsonPayload))

			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(HaveOccurred(),
					"Expected success but got error: %v", err)
				g.Expect(filtered).NotTo(BeEmpty(), "Expected non-empty filtered output")

				// Convert filtered output back to string for substring checks
				filteredStr := string(filtered)
				for _, excl := range tt.expectExclude {
					g.Expect(filteredStr).NotTo(ContainSubstring(excl),
						"Expected %q to be excluded but it was present", excl)
				}
				for _, incl := range tt.expectInclude {
					g.Expect(filteredStr).To(ContainSubstring(incl),
						"Expected %q to be included but it was not found", incl)
				}
			} else {
				g.Expect(err).To(HaveOccurred(),
					"Expected error but got none")
				g.Expect(err.Error()).To(ContainSubstring(tt.expectErrSubstr),
					"Error should contain substring %q", tt.expectErrSubstr)
			}
		})
	}
}
