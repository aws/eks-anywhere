package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

var registryMirrorTests = []struct {
	name                 string
	registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration
	wantFiles            []bootstrapv1.File
}{
	{
		name: "with ca cert",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint:      "1.2.3.4",
			Port:          "443",
			CACertContent: "xyz",
			OCINamespaces: []v1alpha1.OCINamespace{
				{
					Registry:  "public.ecr.aws",
					Namespace: "eks-anywhere",
				},
				{
					Registry:  "783794618700.dkr.ecr.us-west-2.amazonaws.com",
					Namespace: "curated-packages",
				},
				{
					Registry:  "783794618700.dkr.ecr.us-east-1.amazonaws.com",
					Namespace: "curated-packages",
				},
			},
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443/v2/eks-anywhere"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."783794618700.dkr.ecr.*.amazonaws.com"]
    endpoint = ["https://1.2.3.4:443/v2/curated-packages"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    ca_file = "/etc/containerd/certs.d/1.2.3.4:443/ca.crt"`,
			},
			{
				Path:    "/etc/containerd/certs.d/1.2.3.4:443/ca.crt",
				Owner:   "root:root",
				Content: "xyz",
			},
		},
	},
	{
		name: "with insecure skip",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint:           "1.2.3.4",
			Port:               "443",
			InsecureSkipVerify: true,
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."783794618700.dkr.ecr.*.amazonaws.com"]
    endpoint = ["https://1.2.3.4:443"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    insecure_skip_verify = true`,
			},
		},
	},
	{
		name: "with ca cert and insecure skip",
		registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
			Endpoint:           "1.2.3.4",
			Port:               "443",
			CACertContent:      "xyz",
			InsecureSkipVerify: true,
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/containerd/config_append.toml",
				Owner: "root:root",
				Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."783794618700.dkr.ecr.*.amazonaws.com"]
    endpoint = ["https://1.2.3.4:443"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    ca_file = "/etc/containerd/certs.d/1.2.3.4:443/ca.crt"
    insecure_skip_verify = true`,
			},
			{
				Path:    "/etc/containerd/certs.d/1.2.3.4:443/ca.crt",
				Owner:   "root:root",
				Content: "xyz",
			},
		},
	},
}

func TestSetRegistryMirrorInKubeadmControlPlane(t *testing.T) {
	for _, tt := range registryMirrorTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmControlPlane()
			g.Expect(clusterapi.SetRegistryMirrorInKubeadmControlPlane(got, tt.registryMirrorConfig)).To(Succeed())
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.Files = tt.wantFiles
			g.Expect(got).To(Equal(want))
		})
	}
}

func TestSetRegistryMirrorInKubeadmConfigTemplate(t *testing.T) {
	for _, tt := range registryMirrorTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmConfigTemplate()
			g.Expect(clusterapi.SetRegistryMirrorInKubeadmConfigTemplate(got, tt.registryMirrorConfig)).To(Succeed())
			want := wantKubeadmConfigTemplate()
			want.Spec.Template.Spec.Files = tt.wantFiles
			g.Expect(got).To(Equal(want))
		})
	}
}
