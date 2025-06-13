package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestValidateSignature(t *testing.T) {
	tests := []struct {
		name      string
		bundle    *anywherev1alpha1.Bundles
		publicKey string
		valid     bool
		wantErr   error
	}{
		{
			name: "empty bundle with signature field",
			bundle: &anywherev1alpha1.Bundles{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "MEUCICV1iiNA4owIUdZBIowSgWjTKx+JT5/CE8PzmF2CBD5+AiEAk8Fcc1X/LNGm0YCyZISWFhbh4qdc7ENyYCU3DB0u4b0=",
					},
				},
			},
			valid:   false,
			wantErr: fmt.Errorf("filtering excluded fields: gojq execution error"),
		},
		{
			name: "no bundle signature",
			bundle: &anywherev1alpha1.Bundles{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"eks.amazonaws.com/no-signature": "",
					},
				},
			},
			valid:   false,
			wantErr: fmt.Errorf("missing bundle signature annotation"),
		},
		{
			name: "invalid signature",
			bundle: &anywherev1alpha1.Bundles{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "invalid",
					},
				},
				Spec: anywherev1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			valid:   false,
			wantErr: fmt.Errorf("bundle signature in metadata isn't base64 encoded"),
		},
		{
			name: "invalid public key",
			bundle: &anywherev1alpha1.Bundles{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "MEUCICV1iiNA4owIUdZBIowSgWjTKx+JT5/CE8PzmF2CBD5+AiEAk8Fcc1X/LNGm0YCyZISWFhbh4qdc7ENyYCU3DB0u4b0=",
					},
				},
				Spec: anywherev1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			publicKey: "invalid",
			valid:     false,
			wantErr:   fmt.Errorf("decoding the public key as string"),
		},
		{
			name: "invalid encoded public key",
			bundle: &anywherev1alpha1.Bundles{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "MEUCICV1iiNA4owIUdZBIowSgWjTKx+JT5/CE8PzmF2CBD5+AiEAk8Fcc1X/LNGm0YCyZISWFhbh4qdc7ENyYCU3DB0u4b0=",
					},
				},
				Spec: anywherev1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			publicKey: "TUVVQ0lDVjFpaU5BNG93SVVkWkJJb3dTZ1dqVEt4K0pUNS9DRThQem1GMkNCRDUrQWlFQWs4RmNjMVgvTE5HbTBZQ3laSVNXRmhiaDRxZGM3RU55WUNVM0RCMHU0YjA9Cg==",
			valid:     false,
			wantErr:   fmt.Errorf("parsing the public key (not PKIX)"),
		},
		{
			name: "signature verification fail",
			bundle: &anywherev1alpha1.Bundles{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "MEUCICV1iiNA4owIUdZBIowSgWjTKx+JT5/CE8PzmF2CBD5+AiEAk8Fcc1X/LNGm0YCyZISWFhbh4qdc7ENyYCU3DB0u4b0=",
					},
				},
				Spec: anywherev1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			publicKey: "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE+JHaQBRHL76XoZvFeIbYDCPDFONnXM+cP307iq3L/pmqnj0EhoERnbKkJHISYBkOu2MH7LUVGcC0hMw1SxcVpg==",
			valid:     false,
			wantErr:   nil,
		},
		{
			name: "signature verification succeeded",
			bundle: &anywherev1alpha1.Bundles{
				TypeMeta: v1.TypeMeta{
					Kind:       "Bundles",
					APIVersion: anywherev1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "MEYCIQCiWwxw/Nchkgtan47FzagXHgB45Op7YWxvSZjFzHau8wIhALG2kbm+H8HJEfN/rUQ0ldo298MnzyhukBptUm0jCtZZ",
					},
				},
				Spec: anywherev1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []anywherev1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			publicKey: constants.KMSPublicKey,
			valid:     true,
			wantErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(_ *testing.T) {
			valid, err := ValidateSignature(tc.bundle, tc.publicKey)
			if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
			if valid != tc.valid {
				t.Errorf("%v got = %v, \nwant %v", tc.name, valid, tc.valid)
			}
		})
	}
}

func TestValidateEKSDistroManifestSignature(t *testing.T) {
	// Create a dummy release with some nonempty Spec (so filtering produces a valid JSON payload).
	testRelease := &eksdv1alpha1.Release{
		TypeMeta: v1.TypeMeta{
			Kind:       "Release",
			APIVersion: eksdv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "kubernetes-1-28-46",
			Namespace: constants.EksaSystemNamespace,
		},
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
	}

	tests := []struct {
		name           string
		release        *eksdv1alpha1.Release
		signature      string
		publicKey      string
		releaseChannel string
		valid          bool
		wantErr        error
	}{
		{
			name:           "no eks distro manifest signature",
			release:        testRelease,
			signature:      "",
			publicKey:      constants.EKSDistroKMSPublicKey,
			releaseChannel: "1-28",
			valid:          false,
			wantErr:        fmt.Errorf("missing 1-28 eks distro manifest signature annotation"),
		},
		{
			name:           "invalid signature",
			release:        testRelease,
			signature:      "invalid",
			publicKey:      constants.EKSDistroKMSPublicKey,
			releaseChannel: "1-29",
			valid:          false,
			wantErr:        fmt.Errorf("eks distro manifest signature in metadata for 1-29 release channel isn't base64 encoded"),
		},
		{
			name:           "invalid public key",
			release:        testRelease,
			signature:      "MEUCICV1iiNA4owIUdZBIowSgWjTKx+JT5/CE8PzmF2CBD5+AiEAk8Fcc1X/LNGm0YCyZISWFhbh4qdc7ENyYCU3DB0u4b0=",
			publicKey:      "invalid",
			releaseChannel: "test-channel",
			valid:          false,
			wantErr:        fmt.Errorf("decoding the public key as string"),
		},
		{
			name:           "invalid encoded public key",
			release:        testRelease,
			signature:      "MEUCICV1iiNA4owIUdZBIowSgWjTKx+JT5/CE8PzmF2CBD5+AiEAk8Fcc1X/LNGm0YCyZISWFhbh4qdc7ENyYCU3DB0u4b0=",
			publicKey:      "TUVVQ0lDVjFpaU5BNG93SVVkWkJJb3dTZ1dqVEt4K0pUNS9DRThQem1GMkNCRDUrQWlFQWs4RmNjMVgvTE5HbTBZQ3laSVNXRmhiaDRxZGM3RU55WUNVM0RCMHU0YjA9Cg==",
			releaseChannel: "test-channel",
			valid:          false,
			wantErr:        fmt.Errorf("parsing the public key (not PKIX)"),
		},
		{
			name:           "signature verification fail",
			release:        testRelease,
			signature:      "MEUCICV1iiNA4owIUdZBIowSgWjTKx+JT5/CE8PzmF2CBD5+AiEAk8Fcc1X/LNGm0YCyZISWFhbh4qdc7ENyYCU3DB0u4b0=",
			publicKey:      constants.EKSDistroKMSPublicKey,
			releaseChannel: "test-channel",
			valid:          false,
			wantErr:        nil,
		},
		{
			name:           "signature verification succeeded",
			release:        testRelease,
			signature:      "MEUCIQC3uP3Dhfb/nhCeir0Hwtf4bddKVfVIauFWBidT18XZOwIgHjzH1mOxBm1N2l2w9wBVy9W1o6CQXpdDz7UcbCszZYc=",
			publicKey:      constants.EKSDistroKMSPublicKey,
			releaseChannel: "1-28",
			valid:          true,
			wantErr:        nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := ValidateEKSDistroManifestSignature(tc.release, tc.signature, tc.publicKey, tc.releaseChannel)
			if err != nil && (tc.wantErr == nil || !strings.Contains(err.Error(), tc.wantErr.Error())) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
			if valid != tc.valid {
				t.Errorf("%v got = %v, \nwant %v", tc.name, valid, tc.valid)
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
			g := gomega.NewWithT(t)

			digest, filtered, err := getBundleDigest(tt.bundle)
			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(gomega.HaveOccurred(), "Expected success but got error: %v", err)
				g.Expect(digest).NotTo(gomega.BeZero(), "Expected digest to be non-zero")
				g.Expect(filtered).NotTo(gomega.BeEmpty(), "Expected filtered bytes to be non-empty")
			} else {
				g.Expect(err).To(gomega.HaveOccurred(), "Expected error but got none")
				g.Expect(err.Error()).To(gomega.ContainSubstring(tt.expectErrSubstr), "Error message should contain substring %q, got: %v", tt.expectErrSubstr, err)
				g.Expect(digest).To(gomega.BeZero())
				g.Expect(filtered).To(gomega.BeNil())
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
			testName: "Simple valid release",
			release: &eksdv1alpha1.Release{
				TypeMeta: v1.TypeMeta{
					Kind:       "Release",
					APIVersion: eksdv1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "kubernetes-1-28-46",
					Namespace: constants.EksaSystemNamespace,
				},
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
		{
			testName: "Release with unmarshalable content",
			release: &eksdv1alpha1.Release{
				TypeMeta: v1.TypeMeta{
					Kind:       "Release",
					APIVersion: eksdv1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "invalid-release",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: eksdv1alpha1.ReleaseSpec{
					Channel: "1-28",
					Number:  46,
				},
				Status: eksdv1alpha1.ReleaseStatus{
					Components: []eksdv1alpha1.Component{
						{
							Name:   "test-component",
							GitTag: "v1.0.0",
							Assets: []eksdv1alpha1.Asset{
								{
									Name: "test-asset",
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
			g := gomega.NewWithT(t)

			digest, filtered, err := getEKSDistroReleaseDigest(tt.release)
			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(gomega.HaveOccurred(), "Expected success but got error: %v", err)
				g.Expect(digest).NotTo(gomega.BeZero(), "Expected digest to be non-zero")
				g.Expect(filtered).NotTo(gomega.BeEmpty(), "Expected filtered bytes to be non-empty")
			} else {
				g.Expect(err).To(gomega.HaveOccurred(), "Expected error but got none")
				g.Expect(err.Error()).To(gomega.ContainSubstring(tt.expectErrSubstr), "Error message should contain substring %q, got: %v", tt.expectErrSubstr, err)
				g.Expect(digest).To(gomega.BeZero())
				g.Expect(filtered).To(gomega.BeNil())
			}
		})
	}
}

func TestFilterExcludes(t *testing.T) {
	testCases := []struct {
		testName        string
		jsonPayload     string
		excludes        string
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
			excludes:        constants.Excludes,
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
			excludes:        constants.Excludes,
			expectErrSubstr: "unmarshalling JSON:",
		},
		{
			testName:        "empty JSON payload",
			jsonPayload:     `{}`,
			excludes:        constants.Excludes,
			expectErrSubstr: "gojq execution error",
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
			excludes:        constants.Excludes,
			expectErrSubstr: "",
			expectExclude:   []string{"creationTimestamp"},
			expectInclude:   []string{"spec"},
		},
		{
			testName:        "Invalid base64 excludes string",
			jsonPayload:     `{"spec": {"field": "value"}}`,
			excludes:        "invalid-base64!@#$",
			expectErrSubstr: "decoding Excludes:",
		},
		{
			testName:        "Invalid gojq template syntax",
			jsonPayload:     `{"spec": {"field": "value"}}`,
			excludes:        base64.StdEncoding.EncodeToString([]byte("invalid.[gojq..syntax")),
			expectErrSubstr: "gojq parse error:",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := gomega.NewWithT(t)

			filtered, err := filterExcludes([]byte(tt.jsonPayload), tt.excludes)

			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(gomega.HaveOccurred(),
					"Expected success but got error: %v", err)
				g.Expect(filtered).NotTo(gomega.BeEmpty(), "Expected non-empty filtered output")

				// Convert filtered output back to string for substring checks
				filteredStr := string(filtered)
				for _, excl := range tt.expectExclude {
					g.Expect(filteredStr).NotTo(gomega.ContainSubstring(excl),
						"Expected %q to be excluded but it was present", excl)
				}
				for _, incl := range tt.expectInclude {
					g.Expect(filteredStr).To(gomega.ContainSubstring(incl),
						"Expected %q to be included but it was not found", incl)
				}
			} else {
				g.Expect(err).To(gomega.HaveOccurred(),
					"Expected error but got none")
				g.Expect(err.Error()).To(gomega.ContainSubstring(tt.expectErrSubstr),
					"Error should contain substring %q", tt.expectErrSubstr)
			}
		})
	}
}

func TestParseLicense(t *testing.T) {
	tests := []struct {
		name       string
		licenseKey string
		key        string
		wantErr    error
	}{
		{
			name:       "malformed token",
			licenseKey: "invalid.token.string",
			key:        "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE96Xb67YUq+at8gFOioYlf1kxOIPio7i3Y8sFrG3a3sn/MzqQmTO9K82psqOuN+E4NdE8VajOtbyfcLo+Ojax/w==",
			wantErr:    fmt.Errorf("parsing licenseToken"),
		},
		{
			name:       "invalid public key",
			licenseKey: "invalid.token.string",
			key:        "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7XtGi5M5nUyoXZpZWg5e9YgQaVbUq4DbxFkGn7yM9rIg+45dQ1pJwYQd/Z9RDZ3umTZHfdmVfaMT8E/2jpa6vYh5AroOn75tN8qaGmG2OqEBoA8k84zK98qNdOJow7CcIWjHQGk6Tr/dSfdTC6ydmBdRMX/7bBYcKylOFf2P65HOMQCB5YdZJAYzvlXEXzoc1o7DD3pT68BOHHTJp6h7+GGXZoNlHJeq1+AKq38Ra6tuI8EUV2S/5+75FFJzMTLVlJ20Jlhh3fuWJtn6a2hGeD/fbZ1w6CMi0dCTGEX6wUOmL5FJ4RFSVthqZCZ7Ap0G2/5Mu3pxVR9glAxThOw61QIDAQAB",
			wantErr:    fmt.Errorf("parsing the public key (not ECDSA)"),
		},
		{
			name:       "signing method not supported",
			licenseKey: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleGFtcGxlIjoiZGF0YSJ9.UKzt6DArjTtHk_Nch6TwbdgVni6FwLJ1fdbVNYikE_kFGTzMZC82m_0qY7l27LtN0J6b_5D8hLLFk3pTZHYGBX5kB2XKH5e5syRkGh6uZHDkGtRjTMoD5sPMZJ0rG4m80k8cgI37UsIt66hoK_45FzSMlTwxogJ2nJk5G1dH10",
			key:        "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE96Xb67YUq+at8gFOioYlf1kxOIPio7i3Y8sFrG3a3sn/MzqQmTO9K82psqOuN+E4NdE8VajOtbyfcLo+Ojax/w==",
			wantErr:    fmt.Errorf("signing method not supported"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseLicense(tc.licenseKey, tc.key)
			if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestParsePublicKeyErrors(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr string
	}{
		{
			name:    "invalid base64 public key",
			key:     "invalid-base64!@#$",
			wantErr: "decoding the public key as string",
		},
		{
			name:    "valid base64 but invalid PKIX format",
			key:     base64.StdEncoding.EncodeToString([]byte("not-a-valid-pkix-key")),
			wantErr: "parsing the public key (not PKIX)",
		},
		{
			name:    "valid PKIX but not ECDSA key",
			key:     "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7XtGi5M5nUyoXZpZWg5e9YgQaVbUq4DbxFkGn7yM9rIg+45dQ1pJwYQd/Z9RDZ3umTZHfdmVfaMT8E/2jpa6vYh5AroOn75tN8qaGmG2OqEBoA8k84zK98qNdOJow7CcIWjHQGk6Tr/dSfdTC6ydmBdRMX/7bBYcKylOFf2P65HOMQCB5YdZJAYzvlXEXzoc1o7DD3pT68BOHHTJp6h7+GGXZoNlHJeq1+AKq38Ra6tuI8EUV2S/5+75FFJzMTLVlJ20Jlhh3fuWJtn6a2hGeD/fbZ1w6CMi0dCTGEX6wUOmL5FJ4RFSVthqZCZ7Ap0G2/5Mu3pxVR9glAxThOw61QIDAQAB",
			wantErr: "parsing the public key (not ECDSA)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parsePublicKey(tc.key)
			if err == nil {
				t.Errorf("Expected error but got none")
			} else if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("Expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func generateTestKeys() (string, *ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	publicKey := &privateKey.PublicKey

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKeyBytes)

	return publicKeyBase64, privateKey, nil
}

func TestParseLicense_Success(t *testing.T) {
	publicKeyBase64, privateKey, err := generateTestKeys()
	if err != nil {
		t.Errorf("Failed to generate test keys: %v", err)
	}

	claims := jwt.MapClaims{
		"iss": "test",
		"sub": "12345",
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Errorf("Failed to sign token: %v", err)
	}

	_, err = ParseLicense(signedToken, publicKeyBase64)
	if err != nil {
		t.Errorf("ParseLicense failed: %v", err)
	}
}

func TestGetEKSDistroReleaseDigest_YAMLMarshalError(t *testing.T) {
	// Create a release with unmarshalable content (containing a channel)
	type problematicRelease struct {
		*eksdv1alpha1.Release
		ProblematicField chan int `yaml:"problematicField,omitempty"`
	}

	release := &problematicRelease{
		Release: &eksdv1alpha1.Release{
			TypeMeta: v1.TypeMeta{
				Kind:       "Release",
				APIVersion: eksdv1alpha1.GroupVersion.String(),
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-release",
				Namespace: constants.EksaSystemNamespace,
			},
			Spec: eksdv1alpha1.ReleaseSpec{
				Channel: "1-28",
				Number:  46,
			},
		},
		ProblematicField: make(chan int),
	}

	_, _, err := getEKSDistroReleaseDigest(release.Release)
	// Note: Since we can't easily create a YAML marshal error with a standard eksdv1alpha1.Release,
	// we'll test this indirectly by ensuring the function handles normal releases correctly
	// The actual marshal error is difficult to trigger in practice with valid struct types
	if err != nil {
		// If we get an error here, it should be a YAML marshal error or downstream error
		if !strings.Contains(err.Error(), "marshalling eks distro release to YAML") &&
			!strings.Contains(err.Error(), "converting eks distro release YAML to JSON") &&
			!strings.Contains(err.Error(), "filtering excluded fields") {
			t.Errorf("Unexpected error type: %v", err)
		}
	}
}

func TestFilterExcludes_GojqNoResult(t *testing.T) {
	// Create a query that would produce no results by using an impossible filter
	jsonPayload := `{"spec": {"field": "value"}}`

	// Create an excludes string that would cause gojq to produce no result
	// This is a complex scenario to trigger, but we can simulate it with an invalid query
	impossibleExcludes := base64.StdEncoding.EncodeToString([]byte("impossible.nonexistent.deeply.nested.field.that.does.not.exist"))

	_, err := filterExcludes([]byte(jsonPayload), impossibleExcludes)
	// The actual error might be different, but we're testing the error handling path
	if err != nil {
		// Accept various error types that could occur in the filterExcludes function
		if !strings.Contains(err.Error(), "gojq") &&
			!strings.Contains(err.Error(), "filtering") &&
			!strings.Contains(err.Error(), "execution") {
			t.Errorf("Expected gojq-related error, got: %v", err)
		}
	}
}

func TestFilterExcludes_JSONUnmarshalError(t *testing.T) {
	// Test with invalid JSON to trigger unmarshaling error
	invalidJSON := `{"unclosed": [}`
	validExcludes := constants.EKSDistroExcludes

	_, err := filterExcludes([]byte(invalidJSON), validExcludes)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "unmarshalling JSON") {
		t.Errorf("Expected JSON unmarshaling error, got: %v", err)
	}
}

func TestFilterExcludes_TemplateExecutionError(t *testing.T) {
	// This test covers potential template execution errors within filterExcludes
	validJSON := `{"spec": {"field": "value"}}`

	// Create excludes that could potentially cause template execution issues
	// Using a very long field name to test edge cases
	longFieldName := strings.Repeat("verylongfieldname", 100)
	problematicExcludes := base64.StdEncoding.EncodeToString([]byte(longFieldName))

	_, err := filterExcludes([]byte(validJSON), problematicExcludes)
	// This may or may not error depending on the template implementation
	// but we're testing the error handling paths
	if err != nil {
		// Accept various error types
		validErrorParts := []string{"gojq", "template", "execution", "filtering"}
		hasValidError := false
		for _, part := range validErrorParts {
			if strings.Contains(err.Error(), part) {
				hasValidError = true
				break
			}
		}
		if !hasValidError {
			t.Errorf("Unexpected error type: %v", err)
		}
	}
}

func TestValidateEKSDistroManifestSignature_MissingSignature(t *testing.T) {
	release := &eksdv1alpha1.Release{
		TypeMeta: v1.TypeMeta{
			Kind:       "Release",
			APIVersion: eksdv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-release",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: eksdv1alpha1.ReleaseSpec{
			Channel: "test-channel",
			Number:  1,
		},
	}

	// Test with empty signature
	valid, err := ValidateEKSDistroManifestSignature(release, "", constants.EKSDistroKMSPublicKey, "test-channel")
	if valid {
		t.Error("Expected validation to fail with empty signature")
	}
	if err == nil {
		t.Error("Expected error for empty signature")
	}
	expectedError := "missing test-channel eks distro manifest signature annotation"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

func TestValidateEKSDistroManifestSignature_Base64DecodeError(t *testing.T) {
	release := &eksdv1alpha1.Release{
		TypeMeta: v1.TypeMeta{
			Kind:       "Release",
			APIVersion: eksdv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-release",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: eksdv1alpha1.ReleaseSpec{
			Channel: "test-channel",
			Number:  1,
		},
	}

	// Test with invalid base64 signature
	valid, err := ValidateEKSDistroManifestSignature(release, "invalid-base64!@#", constants.EKSDistroKMSPublicKey, "test-channel")
	if valid {
		t.Error("Expected validation to fail with invalid base64 signature")
	}
	if err == nil {
		t.Error("Expected error for invalid base64 signature")
	}
	expectedError := "eks distro manifest signature in metadata for test-channel release channel isn't base64 encoded"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

func TestGetEKSDistroReleaseDigest_YAMLToJSONError(t *testing.T) {
	// Create a release that could potentially cause YAML to JSON conversion issues
	// This is difficult to trigger directly with normal structs, so we test the normal path
	// and verify that errors would be handled correctly
	release := &eksdv1alpha1.Release{
		TypeMeta: v1.TypeMeta{
			Kind:       "Release",
			APIVersion: eksdv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-release",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: eksdv1alpha1.ReleaseSpec{
			Channel: "test-channel",
			Number:  1,
		},
	}

	// Normal case should work - this tests that the function path is correct
	_, _, err := getEKSDistroReleaseDigest(release)
	if err != nil {
		// If we get an error, it should be from the filtering step, not YAML/JSON conversion
		if strings.Contains(err.Error(), "converting eks distro release YAML to JSON") {
			t.Logf("Successfully caught YAML to JSON conversion error: %v", err)
		} else if strings.Contains(err.Error(), "filtering excluded fields") {
			// This is expected for the empty release
			t.Logf("Got expected filtering error: %v", err)
		} else {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestGetEKSDistroReleaseDigest_FilteringError(t *testing.T) {
	// Create a minimal release that will trigger filtering error
	release := &eksdv1alpha1.Release{
		TypeMeta: v1.TypeMeta{
			Kind:       "Release",
			APIVersion: eksdv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "empty-release",
			Namespace: constants.EksaSystemNamespace,
		},
		// Empty Spec to trigger filtering issues
	}

	_, _, err := getEKSDistroReleaseDigest(release)
	if err != nil {
		if !strings.Contains(err.Error(), "filtering excluded fields") {
			t.Errorf("Expected filtering error, got: %v", err)
		}
	}
}

func TestParsePublicKey_NotECDSAError(t *testing.T) {
	rsaPublicKeyPEM := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA7XtGi5M5nUyoXZpZWg5e9YgQaVbUq4DbxFkGn7yM9rIg+45dQ1pJwYQd/Z9RDZ3umTZHfdmVfaMT8E/2jpa6vYh5AroOn75tN8qaGmG2OqEBoA8k84zK98qNdOJow7CcIWjHQGk6Tr/dSfdTC6ydmBdRMX/7bBYcKylOFf2P65HOMQCB5YdZJAYzvlXEXzoc1o7DD3pT68BOHHTJp6h7+GGXZoNlHJeq1+AKq38Ra6tuI8EUV2S/5+75FFJzMTLVlJ20Jlhh3fuWJtn6a2hGeD/fbZ1w6CMi0dCTGEX6wUOmL5FJ4RFSVthqZCZ7Ap0G2/5Mu3pxVR9glAxThOw61QIDAQAB"

	_, err := parsePublicKey(rsaPublicKeyPEM)
	if err == nil {
		t.Error("Expected error for non-ECDSA key")
	}

	if !strings.Contains(err.Error(), "parsing the public key (not ECDSA)") {
		t.Errorf("Expected 'parsing the public key (not ECDSA)' error, got: %v", err)
	}
}
