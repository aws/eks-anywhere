package validations

import (
	"context"
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestValidateExtendedK8sVersionSupport(t *testing.T) {
	ctx := context.Background()
	client := test.NewFakeKubeClient()

	tests := []struct {
		name    string
		cluster anywherev1.Cluster
		bundle  *v1alpha1.Bundles
		client  kubernetes.Client
		wantErr error
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
			wantErr: fmt.Errorf("missing signature annotation"),
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
							KubeVersion:          "1.28",
							EndOfStandardSupport: "2024-31-12",
						},
					},
				},
			},
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
			bundle:  validBundle(),
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
			bundle:  validBundle(),
			wantErr: fmt.Errorf("getting licenseToken"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(_ *testing.T) {
			err := ValidateExtendedK8sVersionSupport(ctx, tc.cluster, tc.bundle, client)
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

func validBundle() *v1alpha1.Bundles {
	return &v1alpha1.Bundles{
		TypeMeta: v1.TypeMeta{
			Kind:       "Bundles",
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				constants.SignatureAnnotation: "MEYCIQC8Fuo81dxibtkvrOFZpbFXZGmJnhLN6bkJjx4YB0fGIQIhAJIxIAl3s26eXqcmS6kAyjDd0NXDlBbM0d/GCHcL2Xoo",
			},
		},
		Spec: v1alpha1.BundlesSpec{
			Number: 1,
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					KubeVersion:          "1.28",
					EndOfStandardSupport: "2024-12-31",
				},
			},
		},
	}
}
