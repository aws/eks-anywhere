package urls_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/utils/urls"
)

func TestReplaceHost(t *testing.T) {
	tests := []struct {
		name   string
		orgURL string
		host   string
		want   string
	}{
		{
			name:   "oci url",
			orgURL: "oci://public.ecr.aws/product/chart",
			host:   "1.2.3.4:443",
			want:   "oci://1.2.3.4:443/product/chart",
		},
		{
			name:   "https url",
			orgURL: "https://public.ecr.aws/product/site",
			host:   "1.2.3.4:443",
			want:   "https://1.2.3.4:443/product/site",
		},
		{
			name:   "container image",
			orgURL: "public.ecr.aws/product/image:tag",
			host:   "1.2.3.4:443",
			want:   "1.2.3.4:443/product/image:tag",
		},
		{
			name:   "empty host",
			orgURL: "public.ecr.aws/product/image:tag",
			want:   "public.ecr.aws/product/image:tag",
		},
		{
			name:   "host contains slashes",
			orgURL: "public.ecr.aws/product/image:tag",
			host:   "1.2.3.4:443/namespace",
			want:   "1.2.3.4:443/namespace/product/image:tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(urls.ReplaceHost(tt.orgURL, tt.host)).To(Equal(tt.want))
		})
	}
}

func TestToAPIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		URL  string
		want string
	}{
		{
			name: "no namespace",
			URL:  "1.2.3.4:443",
			want: "1.2.3.4:443",
		},
		{
			name: "with namespace",
			URL:  "1.2.3.4:443/namespace",
			want: "1.2.3.4:443/v2/namespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(urls.ToAPIEndpoint(tt.URL)).To(Equal(tt.want))
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
			name: "no namespace",
			URLs: map[string]string{
				"public.ecr.aws":                       "1.2.3.4:443",
				"783794618700.dkr.ecr.*.amazonaws.com": "1.2.3.4:443/curated-packages",
			},
			want: map[string]string{
				"public.ecr.aws":                       "1.2.3.4:443",
				"783794618700.dkr.ecr.*.amazonaws.com": "1.2.3.4:443/v2/curated-packages",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := urls.ToAPIEndpoints(tt.URLs)
			g.Expect(len(result)).To(Equal(len(tt.want)))
			for k, v := range tt.want {
				g.Expect(result).Should(HaveKeyWithValue(k, v))
			}
		})
	}
}
