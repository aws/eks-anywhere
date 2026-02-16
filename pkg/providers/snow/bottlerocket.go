package snow

import (
	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const bottlerocketBootstrapImage = "bottlerocket-bootstrap-snow"

func bottlerocketBootstrapSnow(image releasev1.Image) bootstrapv1beta2.BottlerocketBootstrapContainer {
	return bootstrapv1beta2.BottlerocketBootstrapContainer{
		Name:            bottlerocketBootstrapImage,
		ImageRepository: image.Image(),
		ImageTag:        image.Tag(),
		Mode:            "always",
	}
}

func addBottlerocketBootstrapSnowInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, image releasev1.Image) {
	b := bottlerocketBootstrapSnow(image)
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketCustomBootstrapContainers = []bootstrapv1beta2.BottlerocketBootstrapContainer{b}
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketCustomBootstrapContainers = []bootstrapv1beta2.BottlerocketBootstrapContainer{b}
}

func addBottlerocketBootstrapSnowInKubeadmConfigTemplate(kct *bootstrapv1beta2.KubeadmConfigTemplate, image releasev1.Image) {
	kct.Spec.Template.Spec.JoinConfiguration.BottlerocketCustomBootstrapContainers = []bootstrapv1beta2.BottlerocketBootstrapContainer{bottlerocketBootstrapSnow(image)}
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
