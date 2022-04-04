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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(urls.ReplaceHost(tt.orgURL, tt.host)).To(Equal(tt.want))
		})
	}
}
