package nodeupgrader

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	upgradeScript        = "/foo/eksa-upgrades/scripts/upgrade.sh"
	defaultUpgraderImage = "public.ecr.aws/t0n3a9y4/aws/upgrader:v1.28.3-eks-1-28-9"
	controlPlaneLabel    = "node-role.kubernetes.io/control-plane"

	CopierContainerName             = "components-copier"
	ContainerdUpgraderContainerName = "containerd-upgrader"
	CNIPluginsUpgraderContainerName = "cni-plugins-upgrader"
	KubeadmUpgraderContainerName    = "kubeadm-upgrader"
	KubeletUpgradeContainerName     = "kubelet-kubectl-upgrader"
	PostUpgradeContainerName        = "post-upgrade-status"
)

// PodName returns the name of the upgrader pod based on the nodeName
func PodName(nodeName string) string {
	return fmt.Sprintf("%s-node-upgrader", nodeName)
}

func UpgradeFirstControlPlanePod(nodeName, image, kubernetesVersion, etcdVersion string) *corev1.Pod {
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_first_cp", kubernetesVersion, etcdVersion)
	p.Spec.Containers = []corev1.Container{printAndCleanupContainer(image)}

	return p
}

func UpgradeRestControlPlanePod(nodeName, image string) *corev1.Pod {
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_rest_cp")
	p.Spec.Containers = []corev1.Container{printAndCleanupContainer(image)}

	return p
}

func UpgradeWorkerPod(nodeName, image string) *corev1.Pod {
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_worker")
	p.Spec.Containers = []corev1.Container{printAndCleanupContainer(image)}
	return p
}

func upgraderPod(nodeName, image string) *corev1.Pod {
	dirOrCreate := corev1.HostPathDirectoryOrCreate
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      PodName(nodeName),
			Namespace: constants.EksaSystemNamespace,
			Labels: map[string]string{
				"ekd-d-upgrader": "true",
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
		nsenterContainer(image, PostUpgradeContainerName, upgradeScript, "print_status_and_cleanup"),
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
	args = append(args, extraArgs...)

	return corev1.Container{
		Name:    name,
		Image:   image,
		Command: []string{"nsenter"},
		Args:    args,
		SecurityContext: &corev1.SecurityContext{
			Privileged: ptr.Bool(true),
		},
	}
}

func printAndCleanupContainer(image string) corev1.Container {
	return corev1.Container{
		Name:  "done",
		Image: "nginx",
	}
}
