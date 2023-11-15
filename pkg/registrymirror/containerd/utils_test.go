package containerd_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/registrymirror/containerd"
)

func TestToAPIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		URL  string
		want string
	}{
		{
			name: "no namespace",
			URL:  "oci://1.2.3.4:443",
			want: "oci://1.2.3.4:443",
		},
		{
			name: "no namespace",
			URL:  "registry-mirror.test:443",
			want: "registry-mirror.test:443",
		},
		{
			name: "with namespace",
			URL:  "oci://1.2.3.4:443/namespace",
			want: "oci://1.2.3.4:443/v2/namespace",
		},
		{
			name: "with namespace",
			URL:  "registry-mirror.test:443/namespace",
			want: "registry-mirror.test:443/v2/namespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(containerd.ToAPIEndpoint(tt.URL)).To(Equal(tt.want))
		})
	}
}

func TestToAPIEndpoints(t *testing.T) {
	tests := []struct {
		name string
		URLs map[string]string
		want map[string]string
	}{
		{
			name: "mix",
			URLs: map[string]string{
				constants.DefaultCoreEKSARegistry:        "1.2.3.4:443",
				constants.DefaultCuratedPackagesRegistry: "1.2.3.4:443/curated-packages",
			},
			want: map[string]string{
				constants.DefaultCoreEKSARegistry:        "1.2.3.4:443",
				constants.DefaultCuratedPackagesRegistry: "1.2.3.4:443/v2/curated-packages",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := containerd.ToAPIEndpoints(tt.URLs)
			g.Expect(len(result)).To(Equal(len(tt.want)))
			for k, v := range tt.want {
				g.Expect(result).Should(HaveKeyWithValue(k, v))
			}
		})
	}
}
