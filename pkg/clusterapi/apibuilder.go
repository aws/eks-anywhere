package clusterapi

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	clusterKind               = "Cluster"
	kubeadmControlPlaneKind   = "KubeadmControlPlane"
	etcdadmClusterKind        = "EtcdadmCluster"
	kubeadmConfigTemplateKind = "KubeadmConfigTemplate"
	machineDeploymentKind     = "MachineDeployment"
)

var (
	ClusterAPIVersion             = clusterv1.GroupVersion.String()
	kubeadmControlPlaneAPIVersion = controlplanev1.GroupVersion.String()
	bootstrapAPIVersion           = bootstrapv1.GroupVersion.String()
	InfrastructureAPIVersion      = fmt.Sprintf("infrastructure.%s/%s", clusterv1.GroupVersion.Group, clusterv1.GroupVersion.Version)
	etcdClusterAPIVersion         = fmt.Sprintf("etcdcluster.%s/%s", clusterv1.GroupVersion.Group, clusterv1.GroupVersion.Version)
)

func clusterLabels(clusterName string) map[string]string {
	return map[string]string{clusterv1.ClusterLabelName: clusterName}
}

func Cluster(clusterSpec *cluster.Spec) *clusterv1.Cluster {
	clusterName := clusterSpec.GetName()
	cluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ClusterAPIVersion,
			Kind:       clusterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: constants.EksaSystemNamespace,
			Labels:    clusterLabels(clusterName),
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks,
				},
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks,
				},
			},
			ControlPlaneRef: &v1.ObjectReference{
				APIVersion: kubeadmControlPlaneAPIVersion,
				Kind:       kubeadmControlPlaneKind,
				Name:       clusterName,
			},
			InfrastructureRef: &v1.ObjectReference{
				APIVersion: InfrastructureAPIVersion,
				Name:       clusterName,
			},
		},
	}

	if clusterSpec.Spec.ExternalEtcdConfiguration != nil {
		cluster.Spec.ManagedExternalEtcdRef = &v1.ObjectReference{
			APIVersion: etcdClusterAPIVersion,
			Kind:       etcdadmClusterKind,
			Name:       clusterName,
		}
	}

	return cluster
}

func KubeadmControlPlane(clusterSpec *cluster.Spec) *controlplanev1.KubeadmControlPlane {
	kcp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubeadmControlPlaneAPIVersion,
			Kind:       kubeadmControlPlaneKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterSpec.GetName(),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					APIVersion: InfrastructureAPIVersion,
					Name:       clusterSpec.GetName(), // TODO
				},
			},
			KubeadmConfigSpec: bootstrapv1.KubeadmConfigSpec{
				ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
					ImageRepository: clusterSpec.VersionsBundle.KubeDistro.Kubernetes.Repository,
					DNS: bootstrapv1.DNS{
						ImageMeta: bootstrapv1.ImageMeta{
							ImageRepository: clusterSpec.VersionsBundle.KubeDistro.CoreDNS.Repository,
							ImageTag:        clusterSpec.VersionsBundle.KubeDistro.CoreDNS.Tag,
						},
					},
				},
				InitConfiguration: &bootstrapv1.InitConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{},
				},
				PreKubeadmCommands:  []string{},
				PostKubeadmCommands: []string{},
			},
			Version: clusterSpec.VersionsBundle.KubeDistro.Kubernetes.Tag,
		},
	}

	replicas := int32(clusterSpec.Spec.ControlPlaneConfiguration.Count)
	kcp.Spec.Replicas = &replicas

	return kcp
}

func KubeadmConfigTemplate(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) bootstrapv1.KubeadmConfigTemplate {
	kct := bootstrapv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: bootstrapAPIVersion,
			Kind:       kubeadmConfigTemplateKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      workerNodeGroupConfig.Name, // TODO: diff
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: bootstrapv1.KubeadmConfigTemplateSpec{
			Template: bootstrapv1.KubeadmConfigTemplateResource{
				Spec: bootstrapv1.KubeadmConfigSpec{
					JoinConfiguration: &bootstrapv1.JoinConfiguration{
						NodeRegistration: bootstrapv1.NodeRegistrationOptions{},
					},
					PreKubeadmCommands:  []string{},
					PostKubeadmCommands: []string{},
				},
			},
		},
	}
	return kct
}

func MachineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) clusterv1.MachineDeployment {
	clusterName := clusterSpec.GetName()
	md := clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ClusterAPIVersion,
			Kind:       machineDeploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      workerNodeGroupConfig.Name,
			Namespace: constants.EksaSystemNamespace,
			Labels:    clusterLabels(clusterName),
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: clusterName,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{
					Labels: clusterLabels(clusterName),
				},
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							APIVersion: bootstrapAPIVersion,
							Kind:       kubeadmConfigTemplateKind,
							Name:       workerNodeGroupConfig.Name, // TODO: different from vsphere
						},
					},
					ClusterName: clusterName,
					InfrastructureRef: v1.ObjectReference{
						APIVersion: InfrastructureAPIVersion,
						Name:       workerNodeGroupConfig.Name, // TODO: different from vsphere
					},
				},
			},
		},
	}
	replicas := int32(workerNodeGroupConfig.Count)
	md.Spec.Replicas = &replicas

	version := clusterSpec.VersionsBundle.KubeDistro.Kubernetes.Tag
	md.Spec.Template.Spec.Version = &version

	return md
}
