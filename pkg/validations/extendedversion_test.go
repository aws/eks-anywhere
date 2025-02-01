package validations_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestValidateExtendedK8sVersionSupport(t *testing.T) {
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
		bundle  *v1alpha1.Bundles
		client  kubernetes.Client
		wantErr error
	}{
		{
			name:    "No bundle signature",
			cluster: &anywherev1.Cluster{},
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
			name:    "bundle verification succeeded",
			cluster: &anywherev1.Cluster{},
			bundle: &v1alpha1.Bundles{
				TypeMeta: v1.TypeMeta{
					Kind:       "Bundles",
					APIVersion: v1alpha1.GroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						constants.SignatureAnnotation: "MEYCIQCiWwxw/Nchkgtan47FzagXHgB45Op7YWxvSZjFzHau8wIhALG2kbm+H8HJEfN/rUQ0ldo298MnzyhukBptUm0jCtZZ",
					},
				},
				Spec: v1alpha1.BundlesSpec{
					Number: 1,
					VersionsBundles: []v1alpha1.VersionsBundle{
						{
							KubeVersion: "1.31",
						},
					},
				},
			},
			wantErr: fmt.Errorf("missing signature annotation"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(_ *testing.T) {
			err := validations.ValidateExtendedK8sVersionSupport(ctx, tc.cluster, tc.bundle, client)
			if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
