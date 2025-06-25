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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	anywherev1constants "github.com/aws/eks-anywhere/pkg/constants"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
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
			key:             constants.BundlesKmsKey,
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
			key:             constants.BundlesKmsKey,
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
			key:             constants.BundlesKmsKey,
			expectErrSubstr: "",
		},
		{
			testName: "get bundle signature",
			bundle: &anywherev1alpha1.Bundles{
				TypeMeta: v1.TypeMeta{
					Kind:       "Bundles",
					APIVersion: anywherev1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						anywherev1constants.SignatureAnnotation:          "MEUCIBQSvgIhP+DWxZABtdXznRHd3pDoFLeNqi+LcvysJlclAiEAsFCH222IZ1u5hJ0dLdu0NJd2rsJnhKNhxpE+Qg3L7qQ=",
						anywherev1constants.EKSDistroSignatureAnnotation: "",
					},
					Name: "bundles-1",
				},
				Spec: anywherev1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion:          "1.31",
							EndOfStandardSupport: "2026-12-31",
							EksD: anywherev1alpha1.EksDRelease{
								Name:           "test",
								ReleaseChannel: "1-31",
								EksDReleaseUrl: "https://distro.eks.amazonaws.com/kubernetes-1-31/kubernetes-1-31-eks-1.yaml",
							},
						},
					},
				},
			},
			key:             constants.BundlesKmsKey,
			expectErrSubstr: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)

			ctx := context.Background()
			sig, err := GetBundleSignature(ctx, tt.bundle, tt.key)
			if tt.testName == "get bundle signature" {
				fmt.Println(sig)
			}

			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(HaveOccurred(), "Expected no error but got: %v", err)
				g.Expect(sig).NotTo(BeEmpty(), "Expected non-empty signature on success")
			} else {
				g.Expect(err).To(HaveOccurred(), "Expected error but got none")
				g.Expect(err.Error()).To(ContainSubstring(tt.expectErrSubstr), "Error message should contain substring %q, got: %v", tt.expectErrSubstr, err)
				g.Expect(sig).To(BeEmpty(), "Expected empty signature when error occurs")
			}
		})
	}
}

func TestGetEKSDistroManifestSignature(t *testing.T) {
	testCases := []struct {
		testName        string
		bundle          *anywherev1alpha1.Bundles
		key             string
		releaseUrl      string
		expectErrSubstr string
	}{
		{
			testName:        "Invalid release URL",
			bundle:          &anywherev1alpha1.Bundles{},
			key:             constants.EKSDistroManifestKmsKey,
			releaseUrl:      "invalid-test-url",
			expectErrSubstr: "getting eks distro release",
		},
		{
			testName:        "Valid eks distro manifest signature generation",
			bundle:          &anywherev1alpha1.Bundles{},
			key:             constants.EKSDistroManifestKmsKey,
			releaseUrl:      "https://distro.eks.amazonaws.com/kubernetes-1-28/kubernetes-1-28-eks-46.yaml",
			expectErrSubstr: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)

			ctx := context.Background()
			sig, err := GetEKSDistroManifestSignature(ctx, tt.bundle, tt.key, tt.releaseUrl)

			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(HaveOccurred(), "Expected no error but got: %v", err)
				g.Expect(sig).NotTo(BeEmpty(), "Expected non-empty signature on success")
			} else {
				g.Expect(err).To(HaveOccurred(), "Expected error but got none")
				g.Expect(err.Error()).To(ContainSubstring(tt.expectErrSubstr), "Error should contain substring %q, got: %v", tt.expectErrSubstr, err)
				g.Expect(sig).To(BeEmpty(), "Expected empty signature when error occurs")
			}
		})
	}
}

func TestGetBundleDigest(t *testing.T) {
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

			digest, filtered, err := getBundleDigest(tt.bundle)
			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(HaveOccurred(), "Expected success but got error: %v", err)
				g.Expect(digest).NotTo(BeZero(), "Expected non-zero digest")
				g.Expect(filtered).NotTo(BeEmpty(), "Expected non-empty filtered output")
			} else {
				g.Expect(err).To(HaveOccurred(), "Expected error but got none")
				g.Expect(err.Error()).To(ContainSubstring(tt.expectErrSubstr), "Error message should contain substring %q, got: %v", tt.expectErrSubstr, err)
				g.Expect(digest).To(BeZero())
				g.Expect(filtered).To(BeNil())
			}
		})
	}
}

func TestGetEKSDistroReleaseDigest(t *testing.T) {
	testCases := []struct {
		testName        string
		release         *eksdv1alpha1.Release
		expectErrSubstr string
	}{
		{
			testName:        "Simple valid eks distro release",
			release:         &eksdv1alpha1.Release{},
			expectErrSubstr: "",
		},
		{
			testName: "Populated eks distro release",
			release: &eksdv1alpha1.Release{
				Spec: eksdv1alpha1.ReleaseSpec{
					Channel: "1-28",
					Number:  46,
				},
				Status: eksdv1alpha1.ReleaseStatus{
					Components: []eksdv1alpha1.Component{
						{
							Name:   "metrics-server",
							GitTag: "v0.7.2",
							Assets: []eksdv1alpha1.Asset{
								{
									Name: "metrics-server-image",
								},
							},
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

			digest, filtered, err := getEKSDistroReleaseDigest(tt.release)
			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(HaveOccurred(), "Expected success but got error: %v", err)
				g.Expect(digest).NotTo(BeZero(), "Expected non-zero digest")
				g.Expect(filtered).NotTo(BeEmpty(), "Expected non-empty filtered output")
			} else {
				g.Expect(err).To(HaveOccurred(), "Expected error but got none")
				g.Expect(err.Error()).To(ContainSubstring(tt.expectErrSubstr), "Error message should contain substring %q, got: %v", tt.expectErrSubstr, err)
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

			filtered, err := filterExcludes([]byte(tt.jsonPayload), anywherev1constants.Excludes)

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
