package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestSetKubeVipInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.Files = []bootstrapv1.File{
		{
			Path:    "/etc/kubernetes/manifests/kube-vip.yaml",
			Owner:   "root:root",
			Content: test.KubeVipTemplate,
		},
	}

	g.Expect(clusterapi.SetKubeVipInKubeadmControlPlane(got, g.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host, "public.ecr.aws/l0g8r8j6/kube-vip/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433")).To(Succeed())
	g.Expect(got).To(Equal(want))
}
