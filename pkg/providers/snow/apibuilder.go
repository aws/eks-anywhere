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

func CAPICluster(clusterSpec *cluster.Spec, snowCluster *snowv1.AWSSnowCluster, kubeadmControlPlane *controlplanev1.KubeadmControlPlane) *clusterv1.Cluster {
	cluster := clusterapi.Cluster(clusterSpec, snowCluster, kubeadmControlPlane)
	return cluster
}

func KubeadmControlPlane(clusterSpec *cluster.Spec, snowMachineTemplate *snowv1.AWSSnowMachineTemplate) (*controlplanev1.KubeadmControlPlane, error) {
	kcp, err := clusterapi.KubeadmControlPlane(clusterSpec, snowMachineTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed generating KubeadmControlPlane: %v", err)
	}

	// TODO: support unstacked etcd
	stackedEtcdExtraArgs := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local.ExtraArgs
	stackedEtcdExtraArgs["listen-peer-urls"] = "https://0.0.0.0:2380"
	stackedEtcdExtraArgs["listen-client-urls"] = "https://0.0.0.0:2379"

	initConfigKubeletExtraArg := kcp.Spec.KubeadmConfigSpec.InitConfiguration.NodeRegistration.KubeletExtraArgs
	initConfigKubeletExtraArg["provider-id"] = "aws-snow:////'{{ ds.meta_data.instance_id }}'"

	joinConfigKubeletExtraArg := kcp.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.KubeletExtraArgs
	joinConfigKubeletExtraArg["provider-id"] = "aws-snow:////'{{ ds.meta_data.instance_id }}'"

	kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands,
		fmt.Sprintf("/etc/eks/bootstrap.sh %s %s", clusterSpec.VersionsBundle.Snow.KubeVip.VersionedImage(), clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host),
	)

	kcp.Spec.KubeadmConfigSpec.PostKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PostKubeadmCommands,
		fmt.Sprintf("/etc/eks/bootstrap-after.sh %s %s", clusterSpec.VersionsBundle.Snow.KubeVip.VersionedImage(), clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host),
	)

	return kcp, nil
}

func kubeadmConfigTemplate(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) (*bootstrapv1.KubeadmConfigTemplate, error) {
	kct, err := clusterapi.KubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig)
	if err != nil {
		return nil, fmt.Errorf("failed generating KubeadmConfigTemplate: %v", err)
	}

	joinConfigKubeletExtraArg := kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs
	joinConfigKubeletExtraArg["provider-id"] = "aws-snow:////'{{ ds.meta_data.instance_id }}'"

	kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands,
		"/etc/eks/bootstrap.sh",
	)
	return kct, nil
}

func KubeadmConfigTemplates(clusterSpec *cluster.Spec) (map[string]*bootstrapv1.KubeadmConfigTemplate, error) {
	m := make(map[string]*bootstrapv1.KubeadmConfigTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		template, err := kubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, err
		}
		m[workerNodeGroupConfig.Name] = template
	}
	return m, nil
}

func machineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration, kubeadmConfigTemplate *bootstrapv1.KubeadmConfigTemplate, snowMachineTemplate *snowv1.AWSSnowMachineTemplate) clusterv1.MachineDeployment {
	md := clusterapi.MachineDeployment(clusterSpec, workerNodeGroupConfig, kubeadmConfigTemplate, snowMachineTemplate)
	return md
}

func MachineDeployments(clusterSpec *cluster.Spec, kubeadmConfigTemplates map[string]*bootstrapv1.KubeadmConfigTemplate, machineTemplates map[string]*snowv1.AWSSnowMachineTemplate) map[string]*clusterv1.MachineDeployment {
	m := make(map[string]*clusterv1.MachineDeployment, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		deployment := machineDeployment(clusterSpec, workerNodeGroupConfig,
			kubeadmConfigTemplates[workerNodeGroupConfig.Name],
			machineTemplates[workerNodeGroupConfig.MachineGroupRef.Name],
		)
		m[workerNodeGroupConfig.Name] = &deployment
	}
	return m
}

func SnowCluster(clusterSpec *cluster.Spec) *snowv1.AWSSnowCluster {
	cluster := &snowv1.AWSSnowCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterapi.InfrastructureAPIVersion(),
			Kind:       SnowClusterKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterSpec.Cluster.GetName(),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: snowv1.AWSSnowClusterSpec{
			Region: "snow",
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host,
				Port: 6443,
			},
		},
	}
	return cluster
}

func SnowMachineTemplates(clusterSpec *cluster.Spec, machineConfigs map[string]*v1alpha1.SnowMachineConfig) map[string]*snowv1.AWSSnowMachineTemplate {
	m := map[string]*snowv1.AWSSnowMachineTemplate{}

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		smt := SnowMachineTemplate(machineConfigs[workerNodeGroupConfig.MachineGroupRef.Name])
		m[workerNodeGroupConfig.MachineGroupRef.Name] = smt
	}
	return m
}

func SnowMachineTemplate(machineConfig *v1alpha1.SnowMachineConfig) *snowv1.AWSSnowMachineTemplate {
	networkConnector := string(machineConfig.Spec.PhysicalNetworkConnector)
	return &snowv1.AWSSnowMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterapi.InfrastructureAPIVersion(),
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
					InstanceType:       string(machineConfig.Spec.InstanceType),
					SSHKeyName:         &machineConfig.Spec.SshKeyName,
					AMI: snowv1.AWSResourceReference{
						ID: &machineConfig.Spec.AMIID,
					},
					CloudInit: snowv1.CloudInit{
						InsecureSkipSecretsManager: true,
					},
					PhysicalNetworkConnectorType: &networkConnector,
				},
			},
		},
	}
}
