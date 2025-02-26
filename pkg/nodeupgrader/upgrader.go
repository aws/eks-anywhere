package nodeupgrader

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	upgradeBin = "/foo/eksa-upgrades/tools/upgrader"

	// CopierContainerName holds the name of the components copier container.
	CopierContainerName = "components-copier"

	// ContainerdUpgraderContainerName holds the name of the containerd upgrader container.
	ContainerdUpgraderContainerName = "containerd-upgrader"

	// CNIPluginsUpgraderContainerName holds the name of the CNI plugins upgrader container.
	CNIPluginsUpgraderContainerName = "cni-plugins-upgrader"

	// KubeadmUpgraderContainerName holds the name of the kubeadm upgrader container.
	KubeadmUpgraderContainerName = "kubeadm-upgrader"

	// KubeletUpgradeContainerName holds the name of the kubelet/kubectl upgrader container.
	KubeletUpgradeContainerName = "kubelet-kubectl-upgrader"

	// PostUpgradeContainerName holds the name of the post upgrade cleanup/status report container.
	PostUpgradeContainerName = "post-upgrade-status"
)

// PodName returns the name of the upgrader pod based on the nodeName.
func PodName(nodeName string) string {
	return fmt.Sprintf("%s-node-upgrader", nodeName)
}

// UpgradeFirstControlPlanePod returns an upgrader pod that should be deployed on the first control plane node.
func UpgradeFirstControlPlanePod(nodeName, image, kubernetesVersion, etcdVersion string) *corev1.Pod {
	p := upgraderPod(nodeName, image, true)
	p.Spec.InitContainers = containersForUpgrade(true, image, nodeName, "upgrade", "node", "--type", "FirstCP", "--k8sVersion", kubernetesVersion, "--etcdVersion", etcdVersion)
	return p
}

// UpgradeSecondaryControlPlanePod returns an upgrader pod that can be deployed on the remaining control plane nodes.
func UpgradeSecondaryControlPlanePod(nodeName, image, kubernetesVersion string) *corev1.Pod {
	p := upgraderPod(nodeName, image, true)
	p.Spec.InitContainers = containersForUpgrade(true, image, nodeName, "upgrade", "node", "--type", "RestCP", "--k8sVersion", kubernetesVersion)
	return p
}

// UpgradeWorkerPod returns an upgrader pod that can be deployed on worker nodes.
func UpgradeWorkerPod(nodeName, image string) *corev1.Pod {
	p := upgraderPod(nodeName, image, false)
	p.Spec.InitContainers = containersForUpgrade(false, image, nodeName, "upgrade", "node", "--type", "Worker")
	return p
}

func upgraderPod(nodeName, image string, isCP bool) *corev1.Pod {
	volumes := []corev1.Volume{hostComponentsVolume()}
	if isCP {
		volumes = append(volumes, kubeVipVolume())
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PodName(nodeName),
			Namespace: constants.EksaSystemNamespace,
			Labels: map[string]string{
				"eks-d-upgrader": "true",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			HostPID:  true,
			Volumes:  volumes,
			Containers: []corev1.Container{
				nsenterContainer(image, PostUpgradeContainerName, upgradeBin, "upgrade", "status"),
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

func containersForUpgrade(isCP bool, image, nodeName string, kubeadmUpgradeCommand ...string) []corev1.Container {
	return []corev1.Container{
		copierContainer(image, isCP),
		nsenterContainer(image, ContainerdUpgraderContainerName, upgradeBin, "upgrade", "containerd"),
		nsenterContainer(image, CNIPluginsUpgraderContainerName, upgradeBin, "upgrade", "cni-plugins"),
		nsenterContainer(image, KubeadmUpgraderContainerName, append([]string{upgradeBin}, kubeadmUpgradeCommand...)...),
		nsenterContainer(image, KubeletUpgradeContainerName, upgradeBin, "upgrade", "kubelet-kubectl"),
	}
}

func copierContainer(image string, isCP bool) corev1.Container {
	volumeMount := []corev1.VolumeMount{
		{
			Name:      "host-components",
			MountPath: "/usr/host",
		},
	}
	if isCP {
		kubeVipVolMount := corev1.VolumeMount{
			Name:      "kube-vip",
			MountPath: fmt.Sprintf("/eksa-upgrades/%s", constants.KubeVipManifestName),
			SubPath:   constants.KubeVipManifestName,
		}
		volumeMount = append(volumeMount, kubeVipVolMount)
	}
	return corev1.Container{
		Name:         CopierContainerName,
		Image:        image,
		Command:      []string{"cp"},
		Args:         []string{"-r", "/eksa-upgrades", "/usr/host"},
		VolumeMounts: volumeMount,
	}
}

func nsenterContainer(image, name string, extraArgs ...string) corev1.Container {
	args := []string{
		"--target",
		"1",
		"--mount",
		"--uts",
		"--ipc",
		"--net",
	}

	return corev1.Container{
		Name:    name,
		Image:   image,
		Command: []string{"nsenter"},
		Args:    append(args, extraArgs...),
		SecurityContext: &corev1.SecurityContext{
			Privileged: ptr.Bool(true),
		},
	}
}

func hostComponentsVolume() corev1.Volume {
	dirOrCreate := corev1.HostPathDirectoryOrCreate
	return corev1.Volume{
		Name: "host-components",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/foo",
				Type: &dirOrCreate,
			},
		},
	}
}

func kubeVipVolume() corev1.Volume {
	return corev1.Volume{
		Name: "kube-vip",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: constants.KubeVipConfigMapName,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  constants.KubeVipManifestName,
						Path: constants.KubeVipManifestName,
					},
				},
			},
		},
	}
}
