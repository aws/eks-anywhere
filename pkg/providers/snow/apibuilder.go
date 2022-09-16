package snow

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
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
	return clusterapi.Cluster(clusterSpec, snowCluster, kubeadmControlPlane)
}

func KubeadmControlPlane(clusterSpec *cluster.Spec, snowMachineTemplate *snowv1.AWSSnowMachineTemplate) (*controlplanev1.KubeadmControlPlane, error) {
	kcp, err := clusterapi.KubeadmControlPlane(clusterSpec, snowMachineTemplate)
	if err != nil {
		return nil, fmt.Errorf("generating KubeadmControlPlane: %v", err)
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

	if snowMachineTemplate.Spec.Template.Spec.ContainersVolume != nil {
		kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands,
			"/etc/eks/bootstrap-volume.sh",
		)
	}

	kcp.Spec.KubeadmConfigSpec.PostKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PostKubeadmCommands,
		fmt.Sprintf("/etc/eks/bootstrap-after.sh %s %s", clusterSpec.VersionsBundle.Snow.KubeVip.VersionedImage(), clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host),
	)

	if err := clusterapi.SetRegistryMirrorInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration); err != nil {
		return nil, err
	}

	if err := clusterapi.SetProxyConfigInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec); err != nil {
		return nil, err
	}

	clusterapi.CreateContainerdConfigFileInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec)
	clusterapi.RestartContainerdInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec)

	return kcp, nil
}

func KubeadmConfigTemplate(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) (*bootstrapv1.KubeadmConfigTemplate, error) {
	kct, err := clusterapi.KubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig)
	if err != nil {
		return nil, fmt.Errorf("generating KubeadmConfigTemplate: %v", err)
	}

	joinConfigKubeletExtraArg := kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs
	joinConfigKubeletExtraArg["provider-id"] = "aws-snow:////'{{ ds.meta_data.instance_id }}'"

	kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands,
		"/etc/eks/bootstrap.sh",
	)

	if clusterSpec.SnowMachineConfig(workerNodeGroupConfig.MachineGroupRef.Name).Spec.ContainersVolume != nil {
		kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands,
			"/etc/eks/bootstrap-volume.sh",
		)
	}

	if err := clusterapi.SetRegistryMirrorInKubeadmConfigTemplate(kct, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration); err != nil {
		return nil, err
	}

	if err := clusterapi.SetProxyConfigInKubeadmConfigTemplate(kct, clusterSpec.Cluster.Spec); err != nil {
		return nil, err
	}

	clusterapi.CreateContainerdConfigFileInKubeadmConfigTemplate(kct, clusterSpec.Cluster.Spec)
	clusterapi.RestartContainerdInKubeadmConfigTemplate(kct, clusterSpec.Cluster.Spec)

	return kct, nil
}

func machineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration, kubeadmConfigTemplate *bootstrapv1.KubeadmConfigTemplate, snowMachineTemplate *snowv1.AWSSnowMachineTemplate) clusterv1.MachineDeployment {
	return clusterapi.MachineDeployment(clusterSpec, workerNodeGroupConfig, kubeadmConfigTemplate, snowMachineTemplate)
}

func MachineDeployments(clusterSpec *cluster.Spec, kubeadmConfigTemplates map[string]*bootstrapv1.KubeadmConfigTemplate, machineTemplates map[string]*snowv1.AWSSnowMachineTemplate) map[string]*clusterv1.MachineDeployment {
	m := make(map[string]*clusterv1.MachineDeployment, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		deployment := machineDeployment(clusterSpec, workerNodeGroupConfig,
			kubeadmConfigTemplates[workerNodeGroupConfig.Name],
			machineTemplates[workerNodeGroupConfig.Name],
		)
		m[workerNodeGroupConfig.Name] = &deployment
	}
	return m
}

func SnowCluster(clusterSpec *cluster.Spec, credentialsSecret *v1.Secret) *snowv1.AWSSnowCluster {
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
			IdentityRef: &snowv1.AWSSnowIdentityReference{
				Name: credentialsSecret.GetName(),
				Kind: snowv1.AWSSnowIdentityKind(credentialsSecret.GetObjectKind().GroupVersionKind().Kind),
			},
		},
	}
	return cluster
}

func CredentialsSecret(name, namespace string, credsB64, certsB64 []byte) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       string(snowv1.SecretKind),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			v1alpha1.SnowCredentialsKey:  credsB64,
			v1alpha1.SnowCertificatesKey: certsB64,
		},
		Type: v1.SecretTypeOpaque,
	}
}

func CAPASCredentialsSecret(clusterSpec *cluster.Spec, credsB64, certsB64 []byte) *v1.Secret {
	s := CredentialsSecret(CredentialsSecretName(clusterSpec), constants.EksaSystemNamespace, credsB64, certsB64)
	label := map[string]string{
		clusterctlv1.ClusterctlMoveLabelName: "true",
	}
	s.SetLabels(label)
	return s
}

func EksaCredentialsSecret(datacenter *v1alpha1.SnowDatacenterConfig, credsB64, certsB64 []byte) *v1.Secret {
	return CredentialsSecret(datacenter.Spec.IdentityRef.Name, datacenter.GetNamespace(), credsB64, certsB64)
}

func SnowMachineTemplate(name string, machineConfig *v1alpha1.SnowMachineConfig) *snowv1.AWSSnowMachineTemplate {
	networkConnector := string(machineConfig.Spec.PhysicalNetworkConnector)
	return &snowv1.AWSSnowMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterapi.InfrastructureAPIVersion(),
			Kind:       SnowMachineTemplateKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
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
					Devices:                      machineConfig.Spec.Devices,
					ContainersVolume:             machineConfig.Spec.ContainersVolume,
				},
			},
		},
	}
}
