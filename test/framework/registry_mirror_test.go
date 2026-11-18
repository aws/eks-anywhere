package framework

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestPackagesSourceRegistryMirrorPath(t *testing.T) {
	const alias = "x3k6m8v0"
	mirror := func(namespaces ...v1alpha1.OCINamespace) *v1alpha1.RegistryMirrorConfiguration {
		return &v1alpha1.RegistryMirrorConfiguration{Endpoint: "196.18.89.98", Port: "443", OCINamespaces: namespaces}
	}
	tests := []struct {
		name   string
		mirror *v1alpha1.RegistryMirrorConfiguration
		want   string
	}{
		{name: "no registry mirror", mirror: nil, want: alias},
		{name: "no OCI namespaces", mirror: mirror(), want: alias},
		{
			name: "public.ecr.aws namespace differs from the packages registry namespace",
			mirror: mirror(
				v1alpha1.OCINamespace{Registry: "public.ecr.aws", Namespace: "eks-a-test"},
				v1alpha1.OCINamespace{Registry: "docker.io", Namespace: "eks-a-test"},
				v1alpha1.OCINamespace{Registry: "067575901363.dkr.ecr.us-west-2.amazonaws.com", Namespace: alias},
			),
			want: "eks-a-test/" + alias,
		},
		{
			name:   "no public.ecr.aws namespace",
			mirror: mirror(v1alpha1.OCINamespace{Registry: "067575901363.dkr.ecr.us-west-2.amazonaws.com", Namespace: alias}),
			want:   alias,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ClusterE2ETest{ClusterConfig: &cluster.Config{Cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{RegistryMirrorConfiguration: tt.mirror},
			}}}
			if got := e.PackagesSourceRegistryMirrorPath(alias); got != tt.want {
				t.Errorf("PackagesSourceRegistryMirrorPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
