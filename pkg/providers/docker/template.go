package docker

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

func NewDockerTemplateBuilder(now types.NowFunc) providers.TemplateBuilder {
	return &DockerTemplateBuilder{
		now: now,
	}
}

type DockerTemplateBuilder struct {
	now types.NowFunc
}

func (d *DockerTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	values := buildTemplateMapCP(clusterSpec)
	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (d *DockerTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, workloadTemplateNames, kubeadmconfigTemplateNames map[string]string) (content []byte, err error) {
	workerSpecs := make([][]byte, 0, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		values := buildTemplateMapMD(clusterSpec, workerNodeGroupConfiguration)
		values["workloadTemplateName"] = workloadTemplateNames[workerNodeGroupConfiguration.Name]
		values["workloadkubeadmconfigTemplateName"] = kubeadmconfigTemplateNames[workerNodeGroupConfiguration.Name]

		bytes, err := templater.Execute(defaultCAPIConfigMD, values)
		if err != nil {
			return nil, err
		}
		workerSpecs = append(workerSpecs, bytes)
	}

	return templater.AppendYamlResources(workerSpecs...), nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	etcdExtraArgs := clusterapi.SecureEtcdTlsCipherSuitesExtraArgs()
	sharedExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs()
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf)).
		Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(clusterSpec.Cluster.Spec.ControlPlaneConfiguration))
	apiServerExtraArgs := clusterapi.OIDCToExtraArgs(clusterSpec.OIDCConfig).
		Append(clusterapi.AwsIamAuthExtraArgs(clusterSpec.AWSIamConfig)).
		Append(clusterapi.PodIAMAuthExtraArgs(clusterSpec.Cluster.Spec.PodIAMConfig)).
		Append(sharedExtraArgs)
	controllerManagerExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.NodeCIDRMaskExtraArgs(&clusterSpec.Cluster.Spec.ClusterNetwork))

	values := map[string]interface{}{
		"clusterName":                clusterSpec.Cluster.Name,
		"control_plane_replicas":     clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
		"kubernetesRepository":       bundle.KubeDistro.Kubernetes.Repository,
		"kubernetesVersion":          bundle.KubeDistro.Kubernetes.Tag,
		"etcdRepository":             bundle.KubeDistro.Etcd.Repository,
		"etcdVersion":                bundle.KubeDistro.Etcd.Tag,
		"corednsRepository":          bundle.KubeDistro.CoreDNS.Repository,
		"corednsVersion":             bundle.KubeDistro.CoreDNS.Tag,
		"kindNodeImage":              bundle.EksD.KindNode.VersionedImage(),
		"etcdExtraArgs":              etcdExtraArgs.ToPartialYaml(),
		"etcdCipherSuites":           crypto.SecureCipherSuitesString(),
		"apiserverExtraArgs":         apiServerExtraArgs.ToPartialYaml(),
		"controllermanagerExtraArgs": controllerManagerExtraArgs.ToPartialYaml(),
		"schedulerExtraArgs":         sharedExtraArgs.ToPartialYaml(),
		"kubeletExtraArgs":           kubeletExtraArgs.ToPartialYaml(),
		"externalEtcdVersion":        bundle.KubeDistro.EtcdVersion,
		"eksaSystemNamespace":        constants.EksaSystemNamespace,
		"auditPolicy":                common.GetAuditPolicy(),
		"podCidrs":                   clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":               clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks,
		"haproxyImageRepository":     getHAProxyImageRepo(bundle.Haproxy.Image),
		"haproxyImageTag":            bundle.Haproxy.Image.Tag(),
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		values["externalEtcd"] = true
		values["externalEtcdReplicas"] = clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
	}
	if clusterSpec.AWSIamConfig != nil {
		values["awsIamAuth"] = true
	}

	values["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints

	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupConfiguration v1alpha1.WorkerNodeGroupConfiguration) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	kubeletExtraArgs := clusterapi.SecureTlsCipherSuitesExtraArgs().
		Append(clusterapi.WorkerNodeLabelsExtraArgs(workerNodeGroupConfiguration)).
		Append(clusterapi.ResolvConfExtraArgs(clusterSpec.Cluster.Spec.ClusterNetwork.DNS.ResolvConf))

	values := map[string]interface{}{
		"clusterName":           clusterSpec.Cluster.Name,
		"kubernetesVersion":     bundle.KubeDistro.Kubernetes.Tag,
		"kindNodeImage":         bundle.EksD.KindNode.VersionedImage(),
		"eksaSystemNamespace":   constants.EksaSystemNamespace,
		"kubeletExtraArgs":      kubeletExtraArgs.ToPartialYaml(),
		"workerReplicas":        workerNodeGroupConfiguration.Count,
		"workerNodeGroupName":   fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfiguration.Name),
		"workerNodeGroupTaints": workerNodeGroupConfiguration.Taints,
		"autoscalingConfig":     workerNodeGroupConfiguration.AutoScalingConfiguration,
	}

	return values
}
