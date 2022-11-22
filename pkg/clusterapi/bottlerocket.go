package clusterapi

import (
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func bottlerocketBootstrap(image v1alpha1.Image) bootstrapv1.BottlerocketBootstrap {
	return bootstrapv1.BottlerocketBootstrap{
		ImageMeta: bootstrapv1.ImageMeta{
			ImageRepository: image.Image(),
			ImageTag:        image.Tag(),
		},
	}
}

func pause(image v1alpha1.Image) bootstrapv1.Pause {
	return bootstrapv1.Pause{
		ImageMeta: bootstrapv1.ImageMeta{
			ImageRepository: image.Image(),
			ImageTag:        image.Tag(),
		},
	}
}

// SetBottlerocketInKubeadmControlPlane adds bottlerocket bootstrap image metadata in kubeadmControlPlane.
func SetBottlerocketInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, versionsBundle *cluster.VersionsBundle) {
	b := bottlerocketBootstrap(versionsBundle.BottleRocketBootstrap.Bootstrap)
	p := pause(versionsBundle.KubeDistro.Pause)
	kcp.Spec.KubeadmConfigSpec.Format = bootstrapv1.Bottlerocket
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketBootstrap = b
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Pause = p
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketBootstrap = b
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.Pause = p
}

// SetBottlerocketInKubeadmConfigTemplate adds bottlerocket bootstrap image metadata in kubeadmConfigTemplate.
func SetBottlerocketInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, versionsBundle *cluster.VersionsBundle) {
	kct.Spec.Template.Spec.Format = bootstrapv1.Bottlerocket
	kct.Spec.Template.Spec.JoinConfiguration.BottlerocketBootstrap = bottlerocketBootstrap(versionsBundle.BottleRocketBootstrap.Bootstrap)
	kct.Spec.Template.Spec.JoinConfiguration.Pause = pause(versionsBundle.KubeDistro.Pause)
}
