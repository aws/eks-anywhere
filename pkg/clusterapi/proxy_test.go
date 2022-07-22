package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

var proxyTests = []struct {
	name      string
	proxy     *v1alpha1.ProxyConfiguration
	wantFiles []bootstrapv1.File
}{
	{
		name:      "proxy config nil",
		proxy:     nil,
		wantFiles: []bootstrapv1.File{},
	},
	{
		name: "with proxy, pods cidr, service cidr, cp endpoint",
		proxy: &v1alpha1.ProxyConfiguration{
			HttpProxy:  "1.2.3.4:8888",
			HttpsProxy: "1.2.3.4:8888",
			NoProxy: []string{
				"1.2.3.4/0",
				"1.2.3.5/0",
			},
		},
		wantFiles: []bootstrapv1.File{
			{
				Path:  "/etc/systemd/system/containerd.service.d/http-proxy.conf",
				Owner: "root:root",
				Content: `[Service]
Environment="HTTP_PROXY=1.2.3.4:8888"
Environment="HTTPS_PROXY=1.2.3.4:8888"
Environment="NO_PROXY=1.2.3.4/5,1.2.3.4/5,1.2.3.4/0,1.2.3.5/0,localhost,127.0.0.1,.svc,1.2.3.4"`,
			},
		},
	},
}

func TestSetProxyConfigInKubeadmControlPlane(t *testing.T) {
	for _, tt := range proxyTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmControlPlane()
			g.clusterSpec.Cluster.Spec.ProxyConfiguration = tt.proxy
			g.Expect(clusterapi.SetProxyConfigInKubeadmControlPlane(got, g.clusterSpec.Cluster.Spec)).To(Succeed())
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.Files = tt.wantFiles
			g.Expect(got).To(Equal(want))
		})
	}
}

func TestSetProxyConfigInKubeadmConfigTemplate(t *testing.T) {
	for _, tt := range proxyTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmConfigTemplate()
			g.clusterSpec.Cluster.Spec.ProxyConfiguration = tt.proxy
			g.Expect(clusterapi.SetProxyConfigInKubeadmConfigTemplate(got, g.clusterSpec.Cluster.Spec)).To(Succeed())
			want := wantKubeadmConfigTemplate()
			want.Spec.Template.Spec.Files = tt.wantFiles
			g.Expect(got).To(Equal(want))
		})
	}
}

func TestNoProxyDefaults(t *testing.T) {
	g := NewWithT(t)
	want := []string{
		"localhost",
		"127.0.0.1",
		".svc",
	}
	g.Expect(clusterapi.NoProxyDefaults()).To(Equal(want))
}
