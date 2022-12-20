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

// CreateContainerdConfigFileInKubeadmControlPlane adds the prekubeadm command to create containerd config file in kubeadmControlPlane if registry mirror config exists.
func CreateContainerdConfigFileInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, cluster *v1alpha1.Cluster) {
	if cluster.Spec.RegistryMirrorConfiguration != nil {
		kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands, buildContainerdConfigCommands...)
	}
}

// CreateContainerdConfigFileInKubeadmConfigTemplate adds the prekubeadm command to create containerd config file in kubeadmConfigTemplate if registry mirror config exists.
func CreateContainerdConfigFileInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, cluster *v1alpha1.Cluster) {
	if cluster.Spec.RegistryMirrorConfiguration != nil {
		kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, buildContainerdConfigCommands...)
	}
}

// RestartContainerdInKubeadmControlPlane adds the prekubeadm command to restart containerd daemon in kubeadmControlPlane if registry mirror or proxy config exists.
func RestartContainerdInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, cluster *v1alpha1.Cluster) {
	if restartContainerdNeeded(cluster) {
		kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands, restartContainerdCommands...)
	}
}

// RestartContainerdInKubeadmConfigTemplate adds the prekubeadm command to restart containerd daemon in kubeadmConfigTemplate if registry mirror or proxy config exists.
func RestartContainerdInKubeadmConfigTemplate(kct *bootstrapv1.KubeadmConfigTemplate, cluster *v1alpha1.Cluster) {
	if restartContainerdNeeded(cluster) {
		kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands, restartContainerdCommands...)
	}
}

func restartContainerdNeeded(cluster *v1alpha1.Cluster) bool {
	return cluster.Spec.RegistryMirrorConfiguration != nil || cluster.Spec.ProxyConfiguration != nil
}
