package registrymirror_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

func TestFromCluster(t *testing.T) {
	tests := []struct {
		name    string
		cluster *v1alpha1.Cluster
		want    *registrymirror.RegistryMirror
	}{
		{
			name: "with registry mirror",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: &v1alpha1.RegistryMirrorConfiguration{
						Endpoint: "1.2.3.4",
						Port:     "443",
					},
				},
			},
			want: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry: "1.2.3.4:443",
				},
			},
		},
		{
			name: "with registry mirror and namespace",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: &v1alpha1.RegistryMirrorConfiguration{
						Endpoint: "1.2.3.4",
						Port:     "443",
						OCINamespaces: []v1alpha1.OCINamespace{
							{
								Registry:  "public.ecr.aws",
								Namespace: "eks-anywhere",
							},
							{
								Registry:  "783794618700.dkr.ecr.us-west-2.amazonaws.com",
								Namespace: "curated-packages",
							},
						},
					},
				},
			},
			want: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry:        "1.2.3.4:443/eks-anywhere",
					constants.DefaultCuratedPackagesRegistry: "1.2.3.4:443/curated-packages",
				},
			},
		},
		{
			name: "with registry mirror and public.ecr.aws only",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: &v1alpha1.RegistryMirrorConfiguration{
						Endpoint: "1.2.3.4",
						Port:     "443",
						OCINamespaces: []v1alpha1.OCINamespace{
							{
								Registry:  "public.ecr.aws",
								Namespace: "eks-anywhere",
							},
						},
					},
				},
			},
			want: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry: "1.2.3.4:443/eks-anywhere",
				},
			},
		},
		{
			name: "with registry mirror ca and auth",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: &v1alpha1.RegistryMirrorConfiguration{
						Endpoint:      "1.2.3.4",
						Port:          "443",
						Authenticate:  true,
						CACertContent: "xyz",
					},
				},
			},
			want: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry: "1.2.3.4:443",
				},
				Auth:          true,
				CACertContent: "xyz",
			},
		},
		{
			name:    "without registry mirror",
			cluster: &v1alpha1.Cluster{},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := registrymirror.FromCluster(tt.cluster)
			if tt.want == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result.BaseRegistry).To(Equal(tt.want.BaseRegistry))
				g.Expect(len(result.NamespacedRegistryMap)).To(Equal(len(tt.want.NamespacedRegistryMap)))
				for k, v := range tt.want.NamespacedRegistryMap {
					g.Expect(result.NamespacedRegistryMap).Should(HaveKeyWithValue(k, v))
				}
			}
		})
	}
}

func TestFromClusterRegistryMirrorConfiguration(t *testing.T) {
	testCases := []struct {
		testName string
		config   *v1alpha1.RegistryMirrorConfiguration
		want     *registrymirror.RegistryMirror
	}{
		{
			testName: "empty config",
			config:   nil,
			want:     nil,
		},
		{
			testName: "no OCINamespaces",
			config: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "harbor.eksa.demo",
				Port:          "30003",
				OCINamespaces: nil,
				Authenticate:  true,
			},
			want: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry: "harbor.eksa.demo:30003",
				},
				Auth: true,
			},
		},
		{
			testName: "namespace for both eksa and curated packages",
			config: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint: "harbor.eksa.demo",
				Port:     "30003",
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
					{
						Registry:  "783794618700.dkr.ecr.us-west-2.amazonaws.com",
						Namespace: "curated-packages",
					},
				},
			},
			want: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry:        "harbor.eksa.demo:30003/eks-anywhere",
					constants.DefaultCuratedPackagesRegistry: "harbor.eksa.demo:30003/curated-packages",
				},
				Auth: false,
			},
		},
		{
			testName: "namespace for eksa only",
			config: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint: "harbor.eksa.demo",
				Port:     "30003",
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "",
					},
				},
			},
			want: &registrymirror.RegistryMirror{
				BaseRegistry: "harbor.eksa.demo:30003",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry: "harbor.eksa.demo:30003",
				},
				Auth: false,
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			result := registrymirror.FromClusterRegistryMirrorConfiguration(tt.config)
			if tt.want == nil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result.BaseRegistry).To(Equal(tt.want.BaseRegistry))
				g.Expect(len(result.NamespacedRegistryMap)).To(Equal(len(tt.want.NamespacedRegistryMap)))
				for k, v := range tt.want.NamespacedRegistryMap {
					g.Expect(result.NamespacedRegistryMap).Should(HaveKeyWithValue(k, v))
				}
				g.Expect(result.Auth).To(Equal(tt.want.Auth))
			}
		})
	}
}

func TestCoreEKSAMirror(t *testing.T) {
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
					constants.DefaultCoreEKSARegistry: "1.2.3.4:443/eks-anywhere",
				},
			},
			want: "1.2.3.4:443/eks-anywhere",
		},
		{
			testName: "without namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry: "1.2.3.4:443",
				},
			},
			want: "1.2.3.4:443",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.registryMirror.CoreEKSAMirror()).To(Equal(tt.want))
		})
	}
}

func TestCuratedPackagesMirror(t *testing.T) {
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
					constants.DefaultCuratedPackagesRegistry: "1.2.3.4:443/curated-packages",
				},
			},
			want: "1.2.3.4:443/curated-packages",
		},
		{
			testName: "no required namespace",
			registryMirror: &registrymirror.RegistryMirror{
				BaseRegistry: "1.2.3.4:443",
				NamespacedRegistryMap: map[string]string{
					constants.DefaultCoreEKSARegistry: "1.2.3.4:443/eks-anywhere",
				},
			},
			want: "",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.registryMirror.CuratedPackagesMirror()).To(Equal(tt.want))
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
					constants.DefaultCuratedPackagesRegistry: "harbor.eksa.demo:30003/curated-packages",
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
					constants.DefaultCoreEKSARegistry: "harbor.eksa.demo:30003/eks-anywhere",
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
					constants.DefaultCuratedPackagesRegistry: "harbor.eksa.demo:30003/curated-packages",
				},
			},
			URL:  "https://783794618700.dkr.ecr.us-west-2.amazonaws.com/product/site",
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
					constants.DefaultCuratedPackagesRegistry: "harbor.eksa.demo:30003/curated-packages",
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
					constants.DefaultCoreEKSARegistry: "harbor.eksa.demo:30003/eks-anywhere",
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
