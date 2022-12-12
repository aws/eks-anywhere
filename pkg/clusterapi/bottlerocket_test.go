package clusterapi_test

import (
	"testing"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
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

var adminContainer = bootstrapv1.BottlerocketAdmin{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-admin",
		ImageTag:        "0.0.1",
	},
}

var controlContainer = bootstrapv1.BottlerocketControl{
	ImageMeta: bootstrapv1.ImageMeta{
		ImageRepository: "public.ecr.aws/eks-anywhere/bottlerocket-control",
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

func TestSetBottlerocketAdminContainerImageInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketAdmin = adminContainer
	want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketAdmin = adminContainer

	clusterapi.SetBottlerocketAdminContainerImageInKubeadmControlPlane(got, g.clusterSpec.VersionsBundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketControlContainerImageInKubeadmControlPlane(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmControlPlane()
	want := got.DeepCopy()
	want.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketControl = controlContainer
	want.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketControl = controlContainer

	clusterapi.SetBottlerocketControlContainerImageInKubeadmControlPlane(got, g.clusterSpec.VersionsBundle)
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

func TestSetBottlerocketAdminContainerImageInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.JoinConfiguration.BottlerocketAdmin = adminContainer

	clusterapi.SetBottlerocketAdminContainerImageInKubeadmConfigTemplate(got, g.clusterSpec.VersionsBundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketControlContainerImageInKubeadmConfigTemplate(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantKubeadmConfigTemplate()
	want := got.DeepCopy()
	want.Spec.Template.Spec.JoinConfiguration.BottlerocketControl = controlContainer

	clusterapi.SetBottlerocketControlContainerImageInKubeadmConfigTemplate(got, g.clusterSpec.VersionsBundle)
	g.Expect(got).To(Equal(want))
}

func TestSetBottlerocketInEtcdCluster(t *testing.T) {
	g := newApiBuilerTest(t)
	got := wantEtcdCluster()
	want := got.DeepCopy()
	want.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
		BootstrapImage: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
		PauseImage:     "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
	}
	clusterapi.SetBottlerocketInEtcdCluster(got, g.clusterSpec.VersionsBundle)
	g.Expect(got).To(Equal(want))
}
