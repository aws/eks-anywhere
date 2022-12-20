package snow

import (
	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const bottlerocketBootstrapImage = "bottlerocket-bootstrap-snow"

func bottlerocketBootstrapSnow(image releasev1.Image) bootstrapv1.BottlerocketBootstrapContainer {
	return bootstrapv1.BottlerocketBootstrapContainer{
		Name: bottlerocketBootstrapImage,
		ImageMeta: bootstrapv1.ImageMeta{
			ImageRepository: image.Image(),
			ImageTag:        image.Tag(),
		},
		Mode: "always",
	}
}

func addBottlerocketBootstrapSnowInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, image releasev1.Image) {
	b := bottlerocketBootstrapSnow(image)
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketCustomBootstrapContainers = []bootstrapv1.BottlerocketBootstrapContainer{b}
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketCustomBootstrapContainers = []bootstrapv1.BottlerocketBootstrapContainer{b}
}

func addBottlerocketBootstrapSnowInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, image releasev1.Image) {
	kct.Spec.Template.Spec.JoinConfiguration.BottlerocketCustomBootstrapContainers = []bootstrapv1.BottlerocketBootstrapContainer{bottlerocketBootstrapSnow(image)}
}

func addBottlerocketBootstrapSnowInEtcdCluster(etcd *etcdv1.EtcdadmCluster, image releasev1.Image) {
	etcd.Spec.EtcdadmConfigSpec.BottlerocketConfig.CustomBootstrapContainers = []etcdbootstrapv1.BottlerocketBootstrapContainer{
		{
			Name:  bottlerocketBootstrapImage,
			Image: image.VersionedImage(),
			Mode:  "always",
		},
	}
}
