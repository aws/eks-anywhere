package clusterapi

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	clusterKind               = "Cluster"
	kubeadmControlPlaneKind   = "KubeadmControlPlane"
	etcdadmClusterKind        = "EtcdadmCluster"
	kubeadmConfigTemplateKind = "KubeadmConfigTemplate"
	machineDeploymentKind     = "MachineDeployment"
	EKSAClusterLabelName      = "cluster.anywhere.eks.amazonaws.com/cluster-name"
	EKSAClusterLabelNamespace = "cluster.anywhere.eks.amazonaws.com/cluster-namespace"
)

var (
	clusterAPIVersion             = clusterv1.GroupVersion.String()
	kubeadmControlPlaneAPIVersion = controlplanev1.GroupVersion.String()
	bootstrapAPIVersion           = bootstrapv1.GroupVersion.String()
	etcdClusterAPIVersion         = fmt.Sprintf("etcdcluster.%s/%s", clusterv1.GroupVersion.Group, clusterv1.GroupVersion.Version)
)

type APIObject interface {
	runtime.Object
	GetName() string
}

func InfrastructureAPIVersion() string {
	return fmt.Sprintf("infrastructure.%s/%s", clusterv1.GroupVersion.Group, clusterv1.GroupVersion.Version)
}

func eksaClusterLabels(clusterSpec *cluster.Spec) map[string]string {
	return map[string]string{
		EKSAClusterLabelName:      clusterSpec.Cluster.Name,
		EKSAClusterLabelNamespace: clusterSpec.Cluster.Namespace,
	}
}

func capiClusterLabel(clusterSpec *cluster.Spec) map[string]string {
	return map[string]string{
		clusterv1.ClusterLabelName: ClusterName(clusterSpec.Cluster),
	}
}

func capiObjectLabels(clusterSpec *cluster.Spec) map[string]string {
	return mergeLabels(eksaClusterLabels(clusterSpec), capiClusterLabel(clusterSpec))
}

func mergeLabels(labels ...map[string]string) map[string]string {
	size := 0
	for _, l := range labels {
		size += len(l)
	}

	merged := make(map[string]string, size)
	for _, l := range labels {
		for k, v := range l {
			merged[k] = v
		}
	}

	return merged
}

// ClusterName generates the CAPI cluster name for an EKSA Cluster
func ClusterName(cluster *anywherev1.Cluster) string {
	return cluster.Name
}

func Cluster(clusterSpec *cluster.Spec, infrastructureObject, controlPlaneObject APIObject) *clusterv1.Cluster {
	clusterName := clusterSpec.Cluster.GetName()
	cluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterAPIVersion,
			Kind:       clusterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: constants.EksaSystemNamespace,
			Labels:    capiObjectLabels(clusterSpec),
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
				},
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
				},
			},
			ControlPlaneRef: &v1.ObjectReference{
				APIVersion: controlPlaneObject.GetObjectKind().GroupVersionKind().GroupVersion().String(),
				Name:       controlPlaneObject.GetName(),
				Kind:       controlPlaneObject.GetObjectKind().GroupVersionKind().Kind,
			},
			InfrastructureRef: &v1.ObjectReference{
				APIVersion: infrastructureObject.GetObjectKind().GroupVersionKind().GroupVersion().String(),
				Name:       infrastructureObject.GetName(),
				Kind:       infrastructureObject.GetObjectKind().GroupVersionKind().Kind,
			},
		},
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		cluster.Spec.ManagedExternalEtcdRef = &v1.ObjectReference{
			APIVersion: etcdClusterAPIVersion,
			Kind:       etcdadmClusterKind,
			Name:       clusterName,
		}
	}

	return cluster
}

func KubeadmControlPlane(clusterSpec *cluster.Spec, infrastructureObject APIObject) (*controlplanev1.KubeadmControlPlane, error) {
	replicas := int32(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count)

	etcd := bootstrapv1.Etcd{}
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcd.External = &bootstrapv1.ExternalEtcd{
			Endpoints: []string{},
		}
	} else {
		etcd.Local = &bootstrapv1.LocalEtcd{
			ImageMeta: bootstrapv1.ImageMeta{
				ImageRepository: clusterSpec.VersionsBundle.KubeDistro.Etcd.Repository,
				ImageTag:        clusterSpec.VersionsBundle.KubeDistro.Etcd.Tag,
			},
			ExtraArgs: map[string]string{},
		}
	}

	kcp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubeadmControlPlaneAPIVersion,
			Kind:       kubeadmControlPlaneKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KubeadmControlPlaneName(clusterSpec),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					APIVersion: infrastructureObject.GetObjectKind().GroupVersionKind().GroupVersion().String(),
					Kind:       infrastructureObject.GetObjectKind().GroupVersionKind().Kind,
					Name:       infrastructureObject.GetName(),
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
					Etcd: etcd,
					APIServer: bootstrapv1.APIServer{
						ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
							ExtraArgs:    map[string]string{},
							ExtraVolumes: []bootstrapv1.HostPathMount{},
						},
					},
					ControllerManager: bootstrapv1.ControlPlaneComponent{
						ExtraArgs: ControllerManagerArgs(clusterSpec),
					},
				},
				InitConfiguration: &bootstrapv1.InitConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: SecureTlsCipherSuitesExtraArgs().
							Append(ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration)),
						Taints: clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints,
					},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: SecureTlsCipherSuitesExtraArgs().
							Append(ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration)),
						Taints: clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints,
					},
				},
				PreKubeadmCommands:  []string{},
				PostKubeadmCommands: []string{},
				Files:               []bootstrapv1.File{},
			},
			Replicas: &replicas,
			Version:  clusterSpec.VersionsBundle.KubeDistro.Kubernetes.Tag,
		},
	}

	SetIdentityAuthInKubeadmControlPlane(kcp, clusterSpec)

	return kcp, nil
}

func KubeadmConfigTemplate(clusterSpec *cluster.Spec, workerNodeGroupConfig anywherev1.WorkerNodeGroupConfiguration) (*bootstrapv1.KubeadmConfigTemplate, error) {
	kct := &bootstrapv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: bootstrapAPIVersion,
			Kind:       kubeadmConfigTemplateKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultKubeadmConfigTemplateName(clusterSpec, workerNodeGroupConfig),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: bootstrapv1.KubeadmConfigTemplateSpec{
			Template: bootstrapv1.KubeadmConfigTemplateResource{
				Spec: bootstrapv1.KubeadmConfigSpec{
					ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
						ControllerManager: bootstrapv1.ControlPlaneComponent{
							ExtraArgs: map[string]string{},
						},
						APIServer: bootstrapv1.APIServer{
							ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
								ExtraArgs: map[string]string{},
							},
						},
					},
					JoinConfiguration: &bootstrapv1.JoinConfiguration{
						NodeRegistration: bootstrapv1.NodeRegistrationOptions{
							KubeletExtraArgs: WorkerNodeLabelsExtraArgs(workerNodeGroupConfig),
							Taints:           workerNodeGroupConfig.Taints,
						},
					},
					PreKubeadmCommands:  []string{},
					PostKubeadmCommands: []string{},
					Files:               []bootstrapv1.File{},
				},
			},
		},
	}

	return kct, nil
}

func MachineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig anywherev1.WorkerNodeGroupConfiguration, bootstrapObject, infrastructureObject APIObject) clusterv1.MachineDeployment {
	clusterName := clusterSpec.Cluster.GetName()
	replicas := int32(workerNodeGroupConfig.Count)
	version := clusterSpec.VersionsBundle.KubeDistro.Kubernetes.Tag

	return clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterAPIVersion,
			Kind:       machineDeploymentKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      MachineDeploymentName(clusterSpec, workerNodeGroupConfig),
			Namespace: constants.EksaSystemNamespace,
			Labels:    capiObjectLabels(clusterSpec),
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: clusterName,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{
					Labels: capiClusterLabel(clusterSpec),
				},
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							APIVersion: bootstrapObject.GetObjectKind().GroupVersionKind().GroupVersion().String(),
							Kind:       bootstrapObject.GetObjectKind().GroupVersionKind().Kind,
							Name:       bootstrapObject.GetName(),
						},
					},
					ClusterName: clusterName,
					InfrastructureRef: v1.ObjectReference{
						APIVersion: infrastructureObject.GetObjectKind().GroupVersionKind().GroupVersion().String(),
						Kind:       infrastructureObject.GetObjectKind().GroupVersionKind().Kind,
						Name:       infrastructureObject.GetName(),
					},
					Version: &version,
				},
			},
			Replicas: &replicas,
		},
	}
}
