package clusterapi

import (
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

var buildContainerdConfigCommands = []string{
	"cat /etc/containerd/config_append.toml >> /etc/containerd/config.toml",
}

var restartContainerdCommands = []string{
	"sudo systemctl daemon-reload",
	"sudo systemctl restart containerd",
}

func CreateContainerdConfigFileInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, cluster v1alpha1.ClusterSpec) {
	if cluster.RegistryMirrorConfiguration != nil {
		kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands, buildContainerdConfigCommands...)
	}
}

func CreateContainerdConfigFileInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, cluster v1alpha1.ClusterSpec) {
	if cluster.RegistryMirrorConfiguration != nil {
		kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, buildContainerdConfigCommands...)
	}
}

func RestartContainerdInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, cluster v1alpha1.ClusterSpec) {
	if restartContainerdNeeded(cluster) {
		kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands, restartContainerdCommands...)
	}
}

func RestartContainerdInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, cluster v1alpha1.ClusterSpec) {
	if restartContainerdNeeded(cluster) {
		kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, restartContainerdCommands...)
	}
}

func restartContainerdNeeded(cluster v1alpha1.ClusterSpec) bool {
	return cluster.RegistryMirrorConfiguration != nil || cluster.ProxyConfiguration != nil
}
