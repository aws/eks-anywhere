package bundles_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/bundles"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestCharts(t *testing.T) {
	g := NewWithT(t)
	b := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{},
				{},
			},
		},
	}

	g.Expect(bundles.Charts(b)).NotTo(BeEmpty())
}
