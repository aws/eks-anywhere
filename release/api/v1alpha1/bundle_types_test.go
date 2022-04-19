package v1alpha1_test

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
