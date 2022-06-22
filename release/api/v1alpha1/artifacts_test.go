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
			name: "both images",
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
