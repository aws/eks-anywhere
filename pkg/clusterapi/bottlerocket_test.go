package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

var pause = bootstrapv1.Pause{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-distro/kubernetes/pause",
		ImageTag:        "0.0.1",
	},
}

var bootstrap = bootstrapv1.BottlerocketBootstrap{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap",
		ImageTag:        "0.0.1",
	},
}

func TestSetBottlerocketInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.Format = "bottlerocket"
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketBootstrap = bootstrap
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.Pause = pause
	want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketBootstrap = bootstrap
	want.Spec.KubeadmConfigSpec.JoinConfiguration.Pause = pause

	clusterapi.SetBottlerocketInKubeadmControlPlane(got, g.clusterSpec.VersionsBundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.Format = "bottlerocket"
	want.Spec.Template.Spec.JoinConfiguration.BottlerocketBootstrap = bootstrap
	want.Spec.Template.Spec.JoinConfiguration.Pause = pause

	clusterapi.SetBottlerocketInKubeadmConfigTemplate(got, g.clusterSpec.VersionsBundle)
	g.Expect(got).To(Equal(want))
}
