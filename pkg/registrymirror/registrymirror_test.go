package registrymirror_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

func TestRegistryMirrorWithOCINamespace(t *testing.T) {
	testCases := []struct {
		testName       string
		registryMirror *registrymirror.RegistryMirror
		want           string
	}{
		{
			testName: "with namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultRegistry: "1.2.3.4:443/eks-anywhere",
				},
			},
			want: "1.2.3.4:443/eks-anywhere",
		},
		{
			testName: "no required namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultPackageRegistryRegex: "1.2.3.4:443/curated-packages",
				},
			},
			want: "1.2.3.4:443",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.registryMirror.RegistryMirrorWithOCINamespace()).To(Equal(tt.want))
		})
	}
}

func TestRegistryMirrorWithGatedOCINamespace(t *testing.T) {
	testCases := []struct {
		testName       string
		registryMirror *registrymirror.RegistryMirror
		want           string
	}{
		{
			testName: "with namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultPackageRegistryRegex: "1.2.3.4:443/curated-packages",
				},
			},
			want: "1.2.3.4:443/curated-packages",
		},
		{
			testName: "no required namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultRegistry: "1.2.3.4:443/eks-anywhere",
				},
			},
			want: "1.2.3.4:443",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.registryMirror.RegistryMirrorWithGatedOCINamespace()).To(Equal(tt.want))
		})
	}
}

func TestReplaceRegistry(t *testing.T) {
	tests := []struct {
		name           string
		registryMirror *registrymirror.RegistryMirror
		URL            string
		want           string
	}{
		{
			name:           "oci url without registry mirror",
			registryMirror: nil,
			URL:            "oci://public.ecr.aws/product/chart",
			want:           "oci://public.ecr.aws/product/chart",
		},
		{
			name: "oci url without namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultPackageRegistryRegex: "harbor.eksa.demo:30003/curated-packages",
				},
			},
			URL:  "oci://public.ecr.aws/product/chart",
			want: "oci://public.ecr.aws/product/chart",
		},
		{
			name: "oci url with namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultRegistry: "harbor.eksa.demo:30003/eks-anywhere",
				},
			},
			URL:  "oci://public.ecr.aws/product/chart",
			want: "oci://harbor.eksa.demo:30003/eks-anywhere/product/chart",
		},
		{
			name:           "https url without registry mirror",
			registryMirror: nil,
			URL:            "https://public.ecr.aws/product/site",
			want:           "https://public.ecr.aws/product/site",
		},
		{
			name: "https url without namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
			},
			URL:  "https://public.ecr.aws/product/site",
			want: "https://public.ecr.aws/product/site",
		},
		{
			name: "https url with namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultPackageRegistryRegex: "harbor.eksa.demo:30003/curated-packages",
				},
			},
			URL:  "https://783794618700.dkr,ecr.us-west-2.amazonaws.com/product/site",
			want: "https://harbor.eksa.demo:30003/curated-packages/product/site",
		},
		{
			name:           "container image without registry mirror",
			registryMirror: nil,
			URL:            "public.ecr.aws/product/image:tag",
			want:           "public.ecr.aws/product/image:tag",
		},
		{
			name: "container image without namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultPackageRegistryRegex: "harbor.eksa.demo:30003/curated-packages",
				},
			},
			URL:  "public.ecr.aws/product/image:tag",
			want: "public.ecr.aws/product/image:tag",
		},
		{
			name: "container image without namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					registrymirror.DefaultRegistry: "harbor.eksa.demo:30003/eks-anywhere",
				},
			},
			URL:  "public.ecr.aws/product/image:tag",
			want: "harbor.eksa.demo:30003/eks-anywhere/product/image:tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.registryMirror.ReplaceRegistry(tt.URL)).To(Equal(tt.want))
		})
	}
}
