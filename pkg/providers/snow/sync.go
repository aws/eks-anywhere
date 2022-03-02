package snow

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

const (
	SnowClusterKind         = "AWSSnowCluster"
	SnowMachineTemplateKind = "AWSSnowMachineTemplate"
)

func CAPICluster(clusterSpec *cluster.Spec) *clusterv1.Cluster {
	cluster := clusterapi.Cluster(clusterSpec)
	cluster.Spec.InfrastructureRef.Kind = SnowClusterKind
	cluster.Spec.InfrastructureRef.Name = clusterSpec.GetName() // TODO: use capinamegenerator
	return cluster
}

func KubeadmControlPlane(clusterSpec *cluster.Spec) *controlplanev1.KubeadmControlPlane {
	kcp := clusterapi.KubeadmControlPlane(clusterSpec)

	kcp.Spec.MachineTemplate.InfrastructureRef.Kind = "AWSSnowMachineTemplate"

	// TODO: support unstacked etcd
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd = bootstrapv1.Etcd{
		Local: &bootstrapv1.LocalEtcd{
			ImageMeta: bootstrapv1.ImageMeta{
				ImageRepository: clusterSpec.VersionsBundle.KubeDistro.Etcd.Repository,
				ImageTag:        clusterSpec.VersionsBundle.KubeDistro.Etcd.Tag,
			},
			ExtraArgs: map[string]string{
				"listen-peer-urls":   "https://0.0.0.0:2380",
				"listen-client-urls": "https://0.0.0.0:2379",
			},
		},
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer = bootstrapv1.APIServer{
		ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
			ExtraArgs: map[string]string{
				"cloud-provider": "external",
			},
		},
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.ControllerManager = bootstrapv1.ControlPlaneComponent{
		ExtraArgs: map[string]string{
			"cloud-provider": "external",
		},
	}

	kcp.Spec.KubeadmConfigSpec.InitConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{
		"cloud-provider": "external",
		"provider-id":    "aws-snow:////'{{ ds.meta_data.instance_id }}'",
	}

	kcp.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{
		"cloud-provider": "external",
		"provider-id":    "aws-snow:////'{{ ds.meta_data.instance_id }}'",
	}

	kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = []string{
		fmt.Sprintf("/etc/eks/bootstrap.sh %s %s", clusterSpec.VersionsBundle.Snow.KubeVip.VersionedImage(), clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host),
	}

	kcp.Spec.KubeadmConfigSpec.PostKubeadmCommands = []string{
		fmt.Sprintf("/etc/eks/bootstrap-after.sh %s %s", clusterSpec.VersionsBundle.Snow.KubeVip.VersionedImage(), clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host),
	}

	return kcp
}

func kubeadmConfigTemplate(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) bootstrapv1.KubeadmConfigTemplate {
	kct := clusterapi.KubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig)
	kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{
		"cloud-provider": "external",
		"provider-id":    "aws-snow:////'{{ ds.meta_data.instance_id }}'",
	}
	kct.Spec.Template.Spec.PreKubeadmCommands = []string{
		fmt.Sprintf("/etc/eks/bootstrap.sh %s %s", clusterSpec.VersionsBundle.Snow.KubeVip.VersionedImage(), clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host),
	}
	return kct
}

func KubeadmConfigTemplateList(clusterSpec *cluster.Spec) *bootstrapv1.KubeadmConfigTemplateList {
	kctList := &bootstrapv1.KubeadmConfigTemplateList{}
	for _, workerNodeGroupConfig := range clusterSpec.Spec.WorkerNodeGroupConfigurations {
		kctList.Items = append(kctList.Items, kubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig))
	}
	return kctList
}

func machineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) clusterv1.MachineDeployment {
	md := clusterapi.MachineDeployment(clusterSpec, workerNodeGroupConfig)
	md.Spec.Template.Spec.InfrastructureRef.Kind = "AWSSnowMachineTemplate"
	return md
}

func MachineDeploymentList(clusterSpec *cluster.Spec) *clusterv1.MachineDeploymentList {
	mdList := &clusterv1.MachineDeploymentList{}
	for _, workerNodeGroupConfig := range clusterSpec.Spec.WorkerNodeGroupConfigurations {
		mdList.Items = append(mdList.Items, machineDeployment(clusterSpec, workerNodeGroupConfig))
	}
	return mdList
}

func SnowCluster(clusterSpec *cluster.Spec) *snowv1.AWSSnowCluster {
	cluster := &snowv1.AWSSnowCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterapi.InfrastructureAPIVersion,
			Kind:       SnowClusterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterSpec.GetName(),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: snowv1.AWSSnowClusterSpec{
			Region: "snow",
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
				Port: 6443,
			},
		},
	}
	return cluster
}

func SnowMachineTemplatetList(clusterSpec *cluster.Spec, machineConfigs map[string]*v1alpha1.SnowMachineConfig) *snowv1.AWSSnowMachineTemplateList {
	smt := &snowv1.AWSSnowMachineTemplateList{}
	for _, workerNodeGroupConfig := range clusterSpec.Spec.WorkerNodeGroupConfigurations {
		smt.Items = append(smt.Items, *SnowMachineTemplate(machineConfigs[workerNodeGroupConfig.MachineGroupRef.Name]))
	}
	return smt
}

func SnowMachineTemplate(machineConfig *v1alpha1.SnowMachineConfig) *snowv1.AWSSnowMachineTemplate {
	smt := &snowv1.AWSSnowMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterapi.InfrastructureAPIVersion,
			Kind:       SnowMachineTemplateKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineConfig.GetName(), // TODO: capinamegenerator
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: snowv1.AWSSnowMachineTemplateSpec{
			Template: snowv1.AWSSnowMachineTemplateResource{
				Spec: snowv1.AWSSnowMachineSpec{
					IAMInstanceProfile: "control-plane.cluster-api-provider-aws.sigs.k8s.io",
					InstanceType:       machineConfig.Spec.InstanceType,
					SSHKeyName:         &machineConfig.Spec.SshKeyName,
					AMI: snowv1.AWSResourceReference{
						ID: &machineConfig.Spec.AMIID,
					},
					CloudInit: snowv1.CloudInit{
						InsecureSkipSecretsManager: true,
					},
				},
			},
		},
	}
	return smt
}
