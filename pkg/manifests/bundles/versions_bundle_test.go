package bundles_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestVersionsBundleForKubernetesVersion(t *testing.T) {
	versionsBundle121 := releasev1.VersionsBundle{KubeVersion: "1.21"}
	versionsBundle122 := releasev1.VersionsBundle{KubeVersion: "1.22"}
	versionsBundle137 := releasev1.VersionsBundle{
		KubeVersion: "1.37",
		EksD: releasev1.EksDRelease{
			Name:           "kubernetes-1-37-eks-3",
			ReleaseChannel: "1-37",
			KubeVersion:    "v1.37.0",
			EksDReleaseUrl: "https://example.com/kubernetes-1-37-eks-3.yaml",
		},
	}
	b := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				versionsBundle121,
				versionsBundle122,
				versionsBundle137,
			},
		},
	}
	tests := []struct {
		name        string
		kubeVersion string
		want        *releasev1.VersionsBundle
	}{
		{
			name:        "supported version",
			kubeVersion: "1.21",
			want:        &versionsBundle121,
		},
		{
			name:        "Kubernetes 1.37 selects its EKS-D release",
			kubeVersion: "1.37",
			want:        &versionsBundle137,
		},
		{
			name:        "unsupported version",
			kubeVersion: "1.10",
			want:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(bundles.VersionsBundleForKubernetesVersion(b, tt.kubeVersion)).To(Equal(tt.want))
		})
	}
}
