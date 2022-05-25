package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

var kubeadmCommandsTests = []struct {
	name    string
	cluster v1alpha1.ClusterSpec
	want    []string
}{
	{
		name: "registry mirror and proxy config both exist",
		cluster: v1alpha1.ClusterSpec{
			RegistryMirrorConfiguration: nil,
			ProxyConfiguration:          nil,
		},
		want: []string{},
	},
	{
		name: "registry mirror nil",
		cluster: v1alpha1.ClusterSpec{
			RegistryMirrorConfiguration: nil,
			ProxyConfiguration: &v1alpha1.ProxyConfiguration{
				HttpProxy:  "1.2.3.4:8888",
				HttpsProxy: "1.2.3.4:8888",
				NoProxy: []string{
					"1.2.3.4/0",
				},
			},
		},
		want: restartContainerdCommands,
	},
	{
		name: "proxy config nil",
		cluster: v1alpha1.ClusterSpec{
			RegistryMirrorConfiguration: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint: "1.2.3.4",
			},
			ProxyConfiguration: nil,
		},
		want: restartContainerdCommands,
	},
}

func TestRestartContainerdInKubeadmControlPlane(t *testing.T) {
	for _, tt := range kubeadmCommandsTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmControlPlane()
			clusterapi.RestartContainerdInKubeadmControlPlane(got, tt.cluster)
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = tt.want
			g.Expect(got).To(Equal(want))
		})
	}
}

func TestRestartContainerdInKubeadmConfigTemplate(t *testing.T) {
	for _, tt := range kubeadmCommandsTests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmConfigTemplate()
			clusterapi.RestartContainerdInKubeadmConfigTemplate(got, tt.cluster)
			want := wantKubeadmConfigTemplate()
			want.Spec.Template.Spec.PreKubeadmCommands = tt.want
			g.Expect(got).To(Equal(want))
		})
	}
}
