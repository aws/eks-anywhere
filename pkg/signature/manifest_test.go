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
			wantErr: fmt.Errorf("missing signature annotation"),
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
			wantErr: fmt.Errorf("signature in metadata isn't base64 encoded"),
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
			g := gomega.NewWithT(t)

			digest, filtered, err := getDigest(tt.bundle)
			if tt.expectErrSubstr == "" {
				g.Expect(err).NotTo(gomega.HaveOccurred(), "Expected success but got error")
				g.Expect(digest).NotTo(gomega.BeZero(),
					"Expected digest to be non-zero array")
				g.Expect(filtered).NotTo(gomega.BeEmpty(),
					"Expected filtered bytes to be non-empty")
			} else {
				g.Expect(err).To(gomega.HaveOccurred(),
					"Expected error but got none")
				g.Expect(err.Error()).To(gomega.ContainSubstring(tt.expectErrSubstr),
					"Error message should contain substring %q, got: %v",
					tt.expectErrSubstr, err)
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
			testName:        "empty JSON payload",
			jsonPayload:     `{}`,
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
			expectErrSubstr: "",
			expectExclude:   []string{"creationTimestamp"},
			expectInclude:   []string{"spec"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := gomega.NewWithT(t)

			filtered, err := filterExcludes([]byte(tt.jsonPayload))

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
			key:        constants.LicensePublicKey,
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
			key:        constants.LicensePublicKey,
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
