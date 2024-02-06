package snow

import (
	"fmt"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
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
	"github.com/aws/eks-anywhere/pkg/semver"
)

const (
	// SnowClusterKind is the kubernetes object kind for CAPAS Cluster.
	SnowClusterKind = "AWSSnowCluster"
	// SnowMachineTemplateKind is the kubernetes object kind for CAPAS machine template.
	SnowMachineTemplateKind = "AWSSnowMachineTemplate"
	// SnowIPPoolKind is the kubernetes object kind for CAPAS IP pool.
	SnowIPPoolKind                                   = "AWSSnowIPPool"
	ignoreEtcdKubernetesManifestFolderPreflightError = "DirAvailable--etc-kubernetes-manifests"
)

// CAPICluster generates the CAPICluster object for snow provider.
func CAPICluster(clusterSpec *cluster.Spec, snowCluster *snowv1.AWSSnowCluster, kubeadmControlPlane *controlplanev1.KubeadmControlPlane, etcdCluster *etcdv1.EtcdadmCluster) *clusterv1.Cluster {
	return clusterapi.Cluster(clusterSpec, snowCluster, kubeadmControlPlane, etcdCluster)
}

// KubeadmControlPlane generates the kubeadmControlPlane object for snow provider from clusterSpec and snowMachineTemplate.
func KubeadmControlPlane(log logr.Logger, clusterSpec *cluster.Spec, snowMachineTemplate *snowv1.AWSSnowMachineTemplate) (*controlplanev1.KubeadmControlPlane, error) {
	kcp, err := clusterapi.KubeadmControlPlane(clusterSpec, snowMachineTemplate)
	if err != nil {
		return nil, fmt.Errorf("generating KubeadmControlPlane: %v", err)
	}

	versionsBundle := clusterSpec.RootVersionsBundle()

	if err := clusterapi.SetKubeVipInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host, versionsBundle.Snow.KubeVip.VersionedImage()); err != nil {
		return nil, fmt.Errorf("setting kube-vip: %v", err)
	}

	initConfigKubeletExtraArg := kcp.Spec.KubeadmConfigSpec.InitConfiguration.NodeRegistration.KubeletExtraArgs
	initConfigKubeletExtraArg["provider-id"] = "aws-snow:////'{{ ds.meta_data.instance_id }}'"

	joinConfigKubeletExtraArg := kcp.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.KubeletExtraArgs
	joinConfigKubeletExtraArg["provider-id"] = "aws-snow:////'{{ ds.meta_data.instance_id }}'"

	addStackedEtcdExtraArgsInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration)

	machineConfig := clusterSpec.SnowMachineConfig(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)

	osFamily := machineConfig.OSFamily()
	switch osFamily {
	case v1alpha1.Bottlerocket:
		clusterapi.SetProxyConfigInKubeadmControlPlaneForBottlerocket(kcp, clusterSpec.Cluster)
		clusterapi.SetRegistryMirrorInKubeadmControlPlaneForBottlerocket(kcp, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration)
		clusterapi.SetBottlerocketInKubeadmControlPlane(kcp, versionsBundle)
		clusterapi.SetBottlerocketAdminContainerImageInKubeadmControlPlane(kcp, versionsBundle)
		clusterapi.SetBottlerocketControlContainerImageInKubeadmControlPlane(kcp, versionsBundle)
		clusterapi.SetUnstackedEtcdConfigInKubeadmControlPlaneForBottlerocket(kcp, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration)
		addBottlerocketBootstrapSnowInKubeadmControlPlane(kcp, versionsBundle.Snow.BottlerocketBootstrapSnow)
		clusterapi.SetBottlerocketHostConfigInKubeadmControlPlane(kcp, machineConfig.Spec.HostOSConfiguration)

	case v1alpha1.Ubuntu:
		kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands,
			"/etc/eks/bootstrap.sh",
		)
		kubeVersionSemver, err := semver.New(string(clusterSpec.Cluster.Spec.KubernetesVersion) + ".0")
		if err != nil {
			return nil, fmt.Errorf("error converting kubeVersion %v to semver %v", clusterSpec.Cluster.Spec.KubernetesVersion, err)
		}

		kube129Semver, err := semver.New(string(v1alpha1.Kube129) + ".0")
		if err != nil {
			return nil, fmt.Errorf("error converting kubeVersion %v to semver %v", v1alpha1.Kube129, err)
		}

		if kubeVersionSemver.Compare(kube129Semver) != -1 {
			kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(kcp.Spec.KubeadmConfigSpec.PreKubeadmCommands,
				"if [ -f /run/kubeadm/kubeadm.yaml ]; then sed -i 's#path: /etc/kubernetes/admin.conf#path: /etc/kubernetes/super-admin.conf#' /etc/kubernetes/manifests/kube-vip.yaml; fi",
			)
		}

		if err := clusterapi.SetProxyConfigInKubeadmControlPlaneForUbuntu(kcp, clusterSpec.Cluster); err != nil {
			return nil, err
		}
		if err := clusterapi.SetRegistryMirrorInKubeadmControlPlaneForUbuntu(kcp, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration); err != nil {
			return nil, err
		}
		clusterapi.CreateContainerdConfigFileInKubeadmControlPlane(kcp, clusterSpec.Cluster)
		clusterapi.RestartContainerdInKubeadmControlPlane(kcp, clusterSpec.Cluster)
		clusterapi.SetUnstackedEtcdConfigInKubeadmControlPlaneForUbuntu(kcp, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration)
		kcp.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.IgnorePreflightErrors = append(
			kcp.Spec.KubeadmConfigSpec.JoinConfiguration.NodeRegistration.IgnorePreflightErrors,
			ignoreEtcdKubernetesManifestFolderPreflightError,
		)

	default:
		log.Info("Warning: unsupported OS family when setting up KubeadmControlPlane", "OS family", osFamily)
	}

	return kcp, nil
}

// KubeadmConfigTemplate generates the kubeadmConfigTemplate object for snow provider from clusterSpec and workerNodeGroupConfig.
func KubeadmConfigTemplate(log logr.Logger, clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) (*bootstrapv1.KubeadmConfigTemplate, error) {
	kct, err := clusterapi.KubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig)
	if err != nil {
		return nil, fmt.Errorf("generating KubeadmConfigTemplate: %v", err)
	}

	versionsBundle := clusterSpec.RootVersionsBundle()

	joinConfigKubeletExtraArg := kct.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs
	joinConfigKubeletExtraArg["provider-id"] = "aws-snow:////'{{ ds.meta_data.instance_id }}'"

	machineConfig := clusterSpec.SnowMachineConfig(workerNodeGroupConfig.MachineGroupRef.Name)
	osFamily := machineConfig.OSFamily()
	switch osFamily {
	case v1alpha1.Bottlerocket:
		clusterapi.SetProxyConfigInKubeadmConfigTemplateForBottlerocket(kct, clusterSpec.Cluster)
		clusterapi.SetRegistryMirrorInKubeadmConfigTemplateForBottlerocket(kct, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration)
		clusterapi.SetBottlerocketInKubeadmConfigTemplate(kct, versionsBundle)
		clusterapi.SetBottlerocketAdminContainerImageInKubeadmConfigTemplate(kct, versionsBundle)
		clusterapi.SetBottlerocketControlContainerImageInKubeadmConfigTemplate(kct, versionsBundle)
		addBottlerocketBootstrapSnowInKubeadmConfigTemplate(kct, versionsBundle.Snow.BottlerocketBootstrapSnow)
		clusterapi.SetBottlerocketHostConfigInKubeadmConfigTemplate(kct, machineConfig.Spec.HostOSConfiguration)

	case v1alpha1.Ubuntu:
		kct.Spec.Template.Spec.PreKubeadmCommands = append(kct.Spec.Template.Spec.PreKubeadmCommands,
			"/etc/eks/bootstrap.sh",
		)

		if err := clusterapi.SetProxyConfigInKubeadmConfigTemplateForUbuntu(kct, clusterSpec.Cluster); err != nil {
			return nil, err
		}
		if err := clusterapi.SetRegistryMirrorInKubeadmConfigTemplateForUbuntu(kct, clusterSpec.Cluster.Spec.RegistryMirrorConfiguration); err != nil {
			return nil, err
		}
		clusterapi.CreateContainerdConfigFileInKubeadmConfigTemplate(kct, clusterSpec.Cluster)
		clusterapi.RestartContainerdInKubeadmConfigTemplate(kct, clusterSpec.Cluster)

	default:
		log.Info("Warning: unsupported OS family when setting up KubeadmConfigTemplate", "OS family", osFamily)
	}

	return kct, nil
}

func machineDeployment(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration, kubeadmConfigTemplate *bootstrapv1.KubeadmConfigTemplate, snowMachineTemplate *snowv1.AWSSnowMachineTemplate) *clusterv1.MachineDeployment {
	return clusterapi.MachineDeployment(clusterSpec, workerNodeGroupConfig, kubeadmConfigTemplate, snowMachineTemplate)
}

// EtcdadmCluster builds an etcdadmCluster based on an eks-a cluster spec and snowMachineTemplate.
func EtcdadmCluster(log logr.Logger, clusterSpec *cluster.Spec, snowMachineTemplate *snowv1.AWSSnowMachineTemplate) *etcdv1.EtcdadmCluster {
	etcd := clusterapi.EtcdadmCluster(clusterSpec, snowMachineTemplate)

	versionsBundle := clusterSpec.RootVersionsBundle()

	machineConfig := clusterSpec.SnowMachineConfig(clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
	osFamily := machineConfig.OSFamily()
	switch osFamily {
	case v1alpha1.Bottlerocket:
		clusterapi.SetBottlerocketInEtcdCluster(etcd, versionsBundle)
		clusterapi.SetBottlerocketAdminContainerImageInEtcdCluster(etcd, versionsBundle.BottleRocketHostContainers.Admin)
		clusterapi.SetBottlerocketControlContainerImageInEtcdCluster(etcd, versionsBundle.BottleRocketHostContainers.Control)
		addBottlerocketBootstrapSnowInEtcdCluster(etcd, versionsBundle.Snow.BottlerocketBootstrapSnow)
		clusterapi.SetBottlerocketHostConfigInEtcdCluster(etcd, machineConfig.Spec.HostOSConfiguration)

	case v1alpha1.Ubuntu:
		clusterapi.SetUbuntuConfigInEtcdCluster(etcd, versionsBundle, string(*clusterSpec.Cluster.Spec.EksaVersion))
		etcd.Spec.EtcdadmConfigSpec.PreEtcdadmCommands = append(etcd.Spec.EtcdadmConfigSpec.PreEtcdadmCommands,
			"/etc/eks/bootstrap.sh",
		)

	default:
		log.Info("Warning: unsupported OS family when setting up EtcdadmCluster", "OS family", osFamily)
	}

	return etcd
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
		clusterctlv1.ClusterctlMoveLabel: "true",
	}
	s.SetLabels(label)
	return s
}

func EksaCredentialsSecret(datacenter *v1alpha1.SnowDatacenterConfig, credsB64, certsB64 []byte) *v1.Secret {
	return CredentialsSecret(datacenter.Spec.IdentityRef.Name, datacenter.GetNamespace(), credsB64, certsB64)
}

// CAPASIPPools defines a set of CAPAS AWSSnowPool objects.
type CAPASIPPools map[string]*snowv1.AWSSnowIPPool

func (p CAPASIPPools) addPools(dnis []v1alpha1.SnowDirectNetworkInterface, m map[string]*v1alpha1.SnowIPPool) {
	for _, dni := range dnis {
		if dni.IPPoolRef != nil {
			p[dni.IPPoolRef.Name] = toAWSSnowIPPool(m[dni.IPPoolRef.Name])
		}
	}
}

func buildSnowIPPool(pool v1alpha1.IPPool) snowv1.IPPool {
	return snowv1.IPPool{
		IPStart: &pool.IPStart,
		IPEnd:   &pool.IPEnd,
		Gateway: &pool.Gateway,
		Subnet:  &pool.Subnet,
	}
}

func toAWSSnowIPPool(pool *v1alpha1.SnowIPPool) *snowv1.AWSSnowIPPool {
	snowPools := make([]snowv1.IPPool, 0, len(pool.Spec.Pools))
	for _, p := range pool.Spec.Pools {
		snowPools = append(snowPools, buildSnowIPPool(p))
	}

	return &snowv1.AWSSnowIPPool{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterapi.InfrastructureAPIVersion(),
			Kind:       SnowIPPoolKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pool.GetName(),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: snowv1.AWSSnowIPPoolSpec{
			IPPools: snowPools,
		},
	}
}

func buildDNI(dni v1alpha1.SnowDirectNetworkInterface, capasPools CAPASIPPools) snowv1.AWSSnowDirectNetworkInterface {
	var ipPoolRef *v1.ObjectReference
	if dni.IPPoolRef != nil {
		ipPool := capasPools[dni.IPPoolRef.Name]
		ipPoolRef = &v1.ObjectReference{
			Kind: ipPool.Kind,
			Name: ipPool.Name,
		}
	}
	return snowv1.AWSSnowDirectNetworkInterface{
		Index:   dni.Index,
		VlanID:  dni.VlanID,
		DHCP:    dni.DHCP,
		Primary: dni.Primary,
		IPPool:  ipPoolRef,
	}
}

// MachineTemplate builds a snowMachineTemplate based on an eks-a snowMachineConfig and a capasIPPool.
func MachineTemplate(name string, machineConfig *v1alpha1.SnowMachineConfig, capasPools CAPASIPPools) *snowv1.AWSSnowMachineTemplate {
	dnis := make([]snowv1.AWSSnowDirectNetworkInterface, 0, len(machineConfig.Spec.Network.DirectNetworkInterfaces))
	for _, dni := range machineConfig.Spec.Network.DirectNetworkInterfaces {
		dnis = append(dnis, buildDNI(dni, capasPools))
	}

	networkConnector := string(machineConfig.Spec.PhysicalNetworkConnector)

	m := &snowv1.AWSSnowMachineTemplate{
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
					NonRootVolumes:               machineConfig.Spec.NonRootVolumes,
					Network: snowv1.AWSSnowNetwork{
						DirectNetworkInterfaces: dnis,
					},
					OSFamily: (*snowv1.OSFamily)(&machineConfig.Spec.OSFamily),
				},
			},
		},
	}

	if machineConfig.Spec.OSFamily == v1alpha1.Bottlerocket {
		m.Spec.Template.Spec.ImageLookupBaseOS = string(v1alpha1.Bottlerocket)
	}

	return m
}
