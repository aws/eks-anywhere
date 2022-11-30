package clusterapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// SetKubeVipInKubeadmControlPlane appends kube-vip manifest to kubeadmControlPlane's kubeadmConfigSpec files.
func SetKubeVipInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, address, image string) error {
	b, err := yaml.Marshal(kubeVip(address, image))
	if err != nil {
		return fmt.Errorf("marshalling kube-vip pod: %v", err)
	}

	kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, bootstrapv1.File{
		Path:    "/etc/kubernetes/manifests/kube-vip.yaml",
		Owner:   "root:root",
		Content: string(b),
	})

	return nil
}

func kubeVip(address, image string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-vip",
			Namespace: constants.KubeSystemNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "kube-vip",
					Image: image,
					Args:  []string{"manager"},
					Env: []corev1.EnvVar{
						{
							Name:  "vip_arp",
							Value: "true",
						},
						{
							Name:  "port",
							Value: "6443",
						},
						{
							Name:  "vip_cidr",
							Value: "32",
						},
						{
							Name:  "cp_enable",
							Value: "true",
						},
						{
							Name:  "cp_namespace",
							Value: "kube-system",
						},
						{
							Name:  "vip_ddns",
							Value: "false",
						},
						{
							Name:  "vip_leaderelection",
							Value: "true",
						},
						{
							Name:  "vip_leaseduration",
							Value: "15",
						},
						{
							Name:  "vip_renewdeadline",
							Value: "10",
						},
						{
							Name:  "vip_retryperiod",
							Value: "2",
						},
						{
							Name:  "address",
							Value: address,
						},
					},
					ImagePullPolicy: corev1.PullIfNotPresent,
					SecurityContext: &corev1.SecurityContext{
						Capabilities: &corev1.Capabilities{
							Add: []corev1.Capability{
								"NET_ADMIN",
								"NET_RAW",
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "kubeconfig",
							MountPath: "/etc/kubernetes/admin.conf",
						},
					},
				},
			},
			HostNetwork: true,
			Volumes: []corev1.Volume{
				{
					Name: "kubeconfig",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/kubernetes/admin.conf",
						},
					},
				},
			},
		},
	}
}
