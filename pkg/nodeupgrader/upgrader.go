package nodeupgrader

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	upgradeScript = "/foo/eksa-upgrades/scripts/upgrade.sh"

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
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_first_cp", kubernetesVersion, etcdVersion)
	return p
}

// UpgradeSecondaryControlPlanePod returns an upgrader pod that can be deployed on the remaining control plane nodes.
func UpgradeSecondaryControlPlanePod(nodeName, image string) *corev1.Pod {
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_rest_cp")
	return p
}

// UpgradeWorkerPod returns an upgrader pod that can be deployed on worker nodes.
func UpgradeWorkerPod(nodeName, image string) *corev1.Pod {
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_worker")
	return p
}

func upgraderPod(nodeName, image string) *corev1.Pod {
	dirOrCreate := corev1.HostPathDirectoryOrCreate
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
			Volumes: []corev1.Volume{
				{
					Name: "host-components",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/foo",
							Type: &dirOrCreate,
						},
					},
				},
			},
			Containers: []corev1.Container{
				nsenterContainer(image, PostUpgradeContainerName, upgradeScript, "print_status_and_cleanup"),
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
}

func containersForUpgrade(image, nodeName string, kubeadmUpgradeCommand ...string) []corev1.Container {
	return []corev1.Container{
		copierContainer(image),
		nsenterContainer(image, ContainerdUpgraderContainerName, upgradeScript, "upgrade_containerd"),
		nsenterContainer(image, CNIPluginsUpgraderContainerName, upgradeScript, "cni_plugins"),
		nsenterContainer(image, KubeadmUpgraderContainerName, append([]string{upgradeScript}, kubeadmUpgradeCommand...)...),
		nsenterContainer(image, KubeletUpgradeContainerName, upgradeScript, "kubelet_and_kubectl"),
	}
}

func copierContainer(image string) corev1.Container {
	return corev1.Container{
		Name:    CopierContainerName,
		Image:   image,
		Command: []string{"cp"},
		Args:    []string{"-r", "/eksa-upgrades", "/usr/host"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "host-components",
				MountPath: "/usr/host",
			},
		},
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
