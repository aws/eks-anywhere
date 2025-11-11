package validations

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestValidateExtendedK8sVersionSupport(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		cluster     anywherev1.Cluster
		bundle      *v1alpha1.Bundles
		eksdRelease *eksdv1alpha1.Release
		wantErr     error
	}{
		{
			name:    "no bundle signature",
			cluster: anywherev1.Cluster{},
			bundle: &v1alpha1.Bundles{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						"eks.amazonaws.com/no-signature": "",
					},
				},
			},
			wantErr: fmt.Errorf("missing bundle signature annotation"),
		},
		{
			name: "kubernetes version not supported",
			cluster: anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.22",
				},
			},
			bundle:  validBundle(),
			wantErr: fmt.Errorf("getting versions bundle for 1.22 kubernetes version"),
		},
		{
			name: "unsupported EndOfStandardSupport format",
			cluster: anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.28",
				},
			},
			bundle:  bundleWithInvalidDateFormat(),
			wantErr: fmt.Errorf("parsing EndOfStandardSupport field format"),
		},
		{
			name: "missing license token",
			cluster: anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.28",
					LicenseToken:      "",
				},
			},
			bundle: validBundle(),
			eksdRelease: &eksdv1alpha1.Release{
				TypeMeta: v1.TypeMeta{
					Kind:       "Release",
					APIVersion: eksdv1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "kubernetes-1-28-eks-63",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: eksdv1alpha1.ReleaseSpec{
					Channel: "1-28",
					Number:  63,
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
			wantErr: fmt.Errorf("licenseToken is required for extended kubernetes support"),
		},
		{
			name: "invalid licenseKey",
			cluster: anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: "1.28",
					LicenseToken:      "invalid-token",
				},
			},
			bundle: validBundle(),
			eksdRelease: &eksdv1alpha1.Release{
				TypeMeta: v1.TypeMeta{
					Kind:       "Release",
					APIVersion: eksdv1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "kubernetes-1-28-eks-63",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: eksdv1alpha1.ReleaseSpec{
					Channel: "1-28",
					Number:  63,
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
			wantErr: fmt.Errorf("getting licenseToken"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(_ *testing.T) {
			client := test.NewFakeKubeClient()
			if tc.eksdRelease != nil {
				cb := fake.NewClientBuilder()
				cl := cb.WithRuntimeObjects(tc.eksdRelease).Build()
				client = test.NewKubeClient(cl)
			}

			// Use a default empty release manifest if not provided
			releaseManifest := tc.eksdRelease
			if releaseManifest == nil {
				releaseManifest = &eksdv1alpha1.Release{}
			}

			err := ValidateExtendedK8sVersionSupport(ctx, tc.cluster, tc.bundle, releaseManifest, client)
			if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateLicenseKeyIsUnique(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		cluster         *anywherev1.Cluster
		workloadCluster *anywherev1.Cluster
		wantErr         error
	}{
		{
			name: "license key is unique",
			cluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster1",
				},
				Spec: anywherev1.ClusterSpec{
					LicenseToken: "valid-token",
				},
			},
			workloadCluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster2",
				},
				Spec: anywherev1.ClusterSpec{
					LicenseToken: "valid-token1",
				},
			},
			wantErr: nil,
		},
		{
			name: "license key is not unique",
			cluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster1",
				},
				Spec: anywherev1.ClusterSpec{
					LicenseToken: "valid-token",
				},
			},
			workloadCluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster2",
				},
				Spec: anywherev1.ClusterSpec{
					LicenseToken: "valid-token",
				},
			},
			wantErr: fmt.Errorf("license token valid-token is already in use by cluster"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(_ *testing.T) {
			cb := fake.NewClientBuilder()
			cl := cb.WithRuntimeObjects(tc.cluster, tc.workloadCluster).Build()
			client := clientutil.NewKubeClient(cl)

			err := validateLicenseKeyIsUnique(ctx, tc.cluster.Name, tc.cluster.Spec.LicenseToken, client)
			if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateBundleSignature(t *testing.T) {
	tests := []struct {
		name    string
		bundle  *v1alpha1.Bundles
		wantErr string
	}{
		{
			name: "invalid bundle signature",
			bundle: &v1alpha1.Bundles{
				TypeMeta: v1.TypeMeta{
					Kind:       "Bundles",
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "MEYCIQCYJwrDjICgUQImFpJdOLjQlC7OSQutCsqBk+0jUheZTQIhALSj7peTLSTSy9rvNfYwyqbP0fOi3elggWwPcAz89csc",
					},
				},
				Spec: v1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []v1alpha1.VersionsBundle{
						{
							KubeVersion: "1.28",
						},
					},
				},
			},
			wantErr: "signature on the bundle is invalid",
		},
		{
			name:    "valid bundle signature",
			bundle:  validBundle(),
			wantErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateBundleSignature(tc.bundle)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("validateBundleSignature() error = %v, wantErr %v", err, tc.wantErr)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("validateBundleSignature() error = %v, wantErr %v", err, tc.wantErr)
				}
			}
		})
	}
}

func TestValidateEKSDistroManifestSignature(t *testing.T) {
	tests := []struct {
		name           string
		manifest       *eksdv1alpha1.Release
		sig            string
		releaseChannel string
		wantErr        string
	}{
		{
			name: "invalid eks distro manifest signature",
			manifest: &eksdv1alpha1.Release{
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
			},
			sig:            "MEYCIQCYJwrDjICgUQImFpJdOLjQlC7OSQutCsqBk+0jUheZTQIhALSj7peTLSTSy9rvNfYwyqbP0fOi3elggWwPcAz89csc",
			releaseChannel: "1-28",
			wantErr:        "signature on the 1-28 eks distro manifest is invalid",
		},
		{
			name: "valid eks distro manifest signature",
			manifest: &eksdv1alpha1.Release{
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
			},
			sig:            "MEUCIQC3uP3Dhfb/nhCeir0Hwtf4bddKVfVIauFWBidT18XZOwIgHjzH1mOxBm1N2l2w9wBVy9W1o6CQXpdDz7UcbCszZYc=",
			releaseChannel: "1-28",
			wantErr:        "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEKSDistroManifestSignature(tc.manifest, tc.sig, tc.releaseChannel)
			if tc.wantErr == "" {
				if err != nil {
					t.Errorf("validateEKSDistroManifestSignature() error = %v, wantErr %v", err, tc.wantErr)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("validateEKSDistroManifestSignature() error = %v, wantErr %v", err, tc.wantErr)
				}
			}
		})
	}
}

func validBundle() *v1alpha1.Bundles {
	// Load the properly signed bundle from testdata
	// This bundle was signed with AWS KMS and includes all required signatures
	data, err := os.ReadFile("testdata/test-bundle.yaml")
	if err != nil {
		panic(fmt.Sprintf("failed to read test bundle: %v", err))
	}

	bundle := &v1alpha1.Bundles{}
	if err := yaml.Unmarshal(data, bundle); err != nil {
		panic(fmt.Sprintf("failed to unmarshal test bundle: %v", err))
	}

	return bundle
}

func bundleWithInvalidDateFormat() *v1alpha1.Bundles {
	// Load a bundle with an invalid date format (2024-31-12) but valid signature
	// This tests that date parsing validation happens after signature validation
	data, err := os.ReadFile("testdata/bundle-invalid-date.yaml")
	if err != nil {
		panic(fmt.Sprintf("failed to read bundle with invalid date: %v", err))
	}

	bundle := &v1alpha1.Bundles{}
	if err := yaml.Unmarshal(data, bundle); err != nil {
		panic(fmt.Sprintf("failed to unmarshal bundle with invalid date: %v", err))
	}

	return bundle
}
