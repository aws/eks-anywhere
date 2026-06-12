package clusterapi

import (
	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta1"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func bottlerocketBootstrap(image v1alpha1.Image) bootstrapv1beta2.BottlerocketBootstrap {
	return bootstrapv1beta2.BottlerocketBootstrap{
		ImageRepository: image.Image(),
		ImageTag:        image.Tag(),
	}
}

func bottlerocketAdmin(image v1alpha1.Image) bootstrapv1beta2.BottlerocketAdmin {
	return bootstrapv1beta2.BottlerocketAdmin{
		ImageRepository: image.Image(),
		ImageTag:        image.Tag(),
	}
}

func bottlerocketControl(image v1alpha1.Image) bootstrapv1beta2.BottlerocketControl {
	return bootstrapv1beta2.BottlerocketControl{
		ImageRepository: image.Image(),
		ImageTag:        image.Tag(),
	}
}

func pause(image v1alpha1.Image) bootstrapv1beta2.Pause {
	return bootstrapv1beta2.Pause{
		ImageRepository: image.Image(),
		ImageTag:        image.Tag(),
	}
}

func hostConfig(config *anywherev1.HostOSConfiguration) *bootstrapv1beta2.BottlerocketSettings {
	b := &bootstrapv1beta2.BottlerocketSettings{}
	if config.BottlerocketConfiguration.Kernel != nil {
		b.Kernel = &bootstrapv1beta2.BottlerocketKernelSettings{
			SysctlSettings: config.BottlerocketConfiguration.Kernel.SysctlSettings,
		}
	}
	if config.BottlerocketConfiguration.Boot != nil {
		b.Boot = &bootstrapv1beta2.BottlerocketBootSettings{
			BootKernelParameters: config.BottlerocketConfiguration.Boot.BootKernelParameters,
		}
	}
	return b
}

// SetBottlerocketInKubeadmControlPlane adds bottlerocket bootstrap image metadata in kubeadmControlPlane.
func SetBottlerocketInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, versionsBundle *cluster.VersionsBundle) {
	b := bottlerocketBootstrap(versionsBundle.BottleRocketHostContainers.KubeadmBootstrap)
	p := pause(versionsBundle.KubeDistro.Pause)
	kcp.Spec.KubeadmConfigSpec.Format = bootstrapv1beta2.Bottlerocket
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketBootstrap = b
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Pause = p
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketBootstrap = b
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.Pause = p

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes = append(kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager.ExtraVolumes,
		bootstrapv1beta2.HostPathMount{
			HostPath:  "/var/lib/kubeadm/controller-manager.conf",
			MountPath: "/etc/kubernetes/controller-manager.conf",
			Name:      "kubeconfig",
			PathType:  "File",
			ReadOnly:  ptr.Bool(true),
		},
	)

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes = append(kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Scheduler.ExtraVolumes,
		bootstrapv1beta2.HostPathMount{
			HostPath:  "/var/lib/kubeadm/scheduler.conf",
			MountPath: "/etc/kubernetes/scheduler.conf",
			Name:      "kubeconfig",
			PathType:  "File",
			ReadOnly:  ptr.Bool(true),
		},
	)
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.CertificatesDir = "/var/lib/kubeadm/pki"
}

// SetBottlerocketAdminContainerImageInKubeadmControlPlane overrides the default bottlerocket admin container image metadata in kubeadmControlPlane.
func SetBottlerocketAdminContainerImageInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, versionsBundle *cluster.VersionsBundle) {
	b := bottlerocketAdmin(versionsBundle.BottleRocketHostContainers.Admin)
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketAdmin = b
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketAdmin = b
}

// SetBottlerocketControlContainerImageInKubeadmControlPlane overrides the default bottlerocket control container image metadata in kubeadmControlPlane.
func SetBottlerocketControlContainerImageInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, versionsBundle *cluster.VersionsBundle) {
	b := bottlerocketControl(versionsBundle.BottleRocketHostContainers.Control)
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.BottlerocketControl = b
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.BottlerocketControl = b
}

// SetBottlerocketInKubeadmConfigTemplate adds bottlerocket bootstrap image metadata in kubeadmConfigTemplate.
func SetBottlerocketInKubeadmConfigTemplate(kct *bootstrapv1beta2.KubeadmConfigTemplate, versionsBundle *cluster.VersionsBundle) {
	kct.Spec.Template.Spec.Format = bootstrapv1beta2.Bottlerocket
	kct.Spec.Template.Spec.JoinConfiguration.BottlerocketBootstrap = bottlerocketBootstrap(versionsBundle.BottleRocketHostContainers.KubeadmBootstrap)
	kct.Spec.Template.Spec.JoinConfiguration.Pause = pause(versionsBundle.KubeDistro.Pause)
}

// SetBottlerocketAdminContainerImageInKubeadmConfigTemplate overrides the default bottlerocket admin container image metadata in kubeadmConfigTemplate.
func SetBottlerocketAdminContainerImageInKubeadmConfigTemplate(kct *bootstrapv1beta2.KubeadmConfigTemplate, versionsBundle *cluster.VersionsBundle) {
	kct.Spec.Template.Spec.JoinConfiguration.BottlerocketAdmin = bottlerocketAdmin(versionsBundle.BottleRocketHostContainers.Admin)
}

// SetBottlerocketControlContainerImageInKubeadmConfigTemplate overrides the default bottlerocket control container image metadata in kubeadmConfigTemplate.
func SetBottlerocketControlContainerImageInKubeadmConfigTemplate(kct *bootstrapv1beta2.KubeadmConfigTemplate, versionsBundle *cluster.VersionsBundle) {
	kct.Spec.Template.Spec.JoinConfiguration.BottlerocketControl = bottlerocketControl(versionsBundle.BottleRocketHostContainers.Control)
}

// SetBottlerocketHostConfigInKubeadmControlPlane sets bottlerocket specific kernel settings in kubeadmControlPlane.
func SetBottlerocketHostConfigInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, hostOSConfig *anywherev1.HostOSConfiguration) {
	if hostOSConfig == nil || hostOSConfig.BottlerocketConfiguration == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Bottlerocket = hostConfig(hostOSConfig)
	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.Bottlerocket = hostConfig(hostOSConfig)
}

// SetBottlerocketHostConfigInKubeadmConfigTemplate sets bottlerocket specific kernel settings in kubeadmConfigTemplate.
func SetBottlerocketHostConfigInKubeadmConfigTemplate(kct *bootstrapv1beta2.KubeadmConfigTemplate, hostOSConfig *anywherev1.HostOSConfiguration) {
	if hostOSConfig == nil || hostOSConfig.BottlerocketConfiguration == nil {
		return
	}

	kct.Spec.Template.Spec.JoinConfiguration.Bottlerocket = hostConfig(hostOSConfig)
}

// SetBottlerocketInEtcdCluster adds bottlerocket config in etcdadmCluster.
func SetBottlerocketInEtcdCluster(etcd *etcdv1.EtcdadmCluster, versionsBundle *cluster.VersionsBundle) {
	etcd.Spec.EtcdadmConfigSpec.Format = etcdbootstrapv1.Format(anywherev1.Bottlerocket)
	etcd.Spec.EtcdadmConfigSpec.BottlerocketConfig = &etcdbootstrapv1.BottlerocketConfig{
		EtcdImage:      versionsBundle.KubeDistro.EtcdImage.VersionedImage(),
		BootstrapImage: versionsBundle.BottleRocketHostContainers.KubeadmBootstrap.VersionedImage(),
		PauseImage:     versionsBundle.KubeDistro.Pause.VersionedImage(),
	}
}

// SetBottlerocketAdminContainerImageInEtcdCluster overrides the default bottlerocket admin container image metadata in etcdadmCluster.
func SetBottlerocketAdminContainerImageInEtcdCluster(etcd *etcdv1.EtcdadmCluster, adminImage v1alpha1.Image) {
	etcd.Spec.EtcdadmConfigSpec.BottlerocketConfig.AdminImage = adminImage.VersionedImage()
}

// SetBottlerocketControlContainerImageInEtcdCluster overrides the default bottlerocket control container image metadata in etcdadmCluster.
func SetBottlerocketControlContainerImageInEtcdCluster(etcd *etcdv1.EtcdadmCluster, controlImage v1alpha1.Image) {
	etcd.Spec.EtcdadmConfigSpec.BottlerocketConfig.ControlImage = controlImage.VersionedImage()
}

// SetBottlerocketHostConfigInEtcdCluster sets bottlerocket specific kernel settings in etcdadmCluster.
func SetBottlerocketHostConfigInEtcdCluster(etcd *etcdv1.EtcdadmCluster, hostOSConfig *anywherev1.HostOSConfiguration) {
	if hostOSConfig == nil || hostOSConfig.BottlerocketConfiguration == nil {
		return
	}

	if hostOSConfig.BottlerocketConfiguration.Kernel != nil {
		etcd.Spec.EtcdadmConfigSpec.BottlerocketConfig.Kernel = &bootstrapv1.BottlerocketKernelSettings{
			SysctlSettings: hostOSConfig.BottlerocketConfiguration.Kernel.SysctlSettings,
		}
	}
	if hostOSConfig.BottlerocketConfiguration.Boot != nil {
		etcd.Spec.EtcdadmConfigSpec.BottlerocketConfig.Boot = &bootstrapv1.BottlerocketBootSettings{
			BootKernelParameters: hostOSConfig.BottlerocketConfiguration.Boot.BootKernelParameters,
		}
	}
}
