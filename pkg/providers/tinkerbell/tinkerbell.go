package tinkerbell

import (
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

//go:embed config/template-cp.yaml
var defaultCAPIConfigCP string

//go:embed config/template-md.yaml
var defaultClusterConfigMD string

type tinkerbellProvider struct {
	hardwareConfig string
}

type TinkerbellTemplateBuilder struct {
	controlPlaneMachineSpec    *v1alpha1.TinkerbellMachineConfigSpec
	workerNodeGroupMachineSpec *v1alpha1.TinkerbellMachineConfigSpec
	etcdMachineSpec            *v1alpha1.TinkerbellMachineConfigSpec
	now                        types.NowFunc
}

func NewProvider(hardwareConfig string) *tinkerbellProvider {
	return &tinkerbellProvider{
		hardwareConfig: hardwareConfig,
	}
}

func NewTinkerbellTemplateBuilder(controlPlaneMachineSpec, workerNodeGroupMachineSpec, etcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec, now types.NowFunc) providers.TemplateBuilder {
	return &TinkerbellTemplateBuilder{
		controlPlaneMachineSpec:    controlPlaneMachineSpec,
		workerNodeGroupMachineSpec: workerNodeGroupMachineSpec,
		etcdMachineSpec:            etcdMachineSpec,
		now:                        now,
	}
}

func (vs *TinkerbellTemplateBuilder) WorkerMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-worker-node-template-%d", clusterName, t)
}

func (vs *TinkerbellTemplateBuilder) CPMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func (vs *TinkerbellTemplateBuilder) EtcdMachineTemplateName(clusterName string) string {
	t := vs.now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func (vs *TinkerbellTemplateBuilder) GenerateCAPISpecControlPlane(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	values := buildTemplateMapCP(clusterSpec, *vs.controlPlaneMachineSpec)
	for _, buildOption := range buildOptions {
		buildOption(values)
	}
	bytes, err := templater.Execute(defaultCAPIConfigCP, values)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (vs *TinkerbellTemplateBuilder) GenerateCAPISpecWorkers(clusterSpec *cluster.Spec, buildOptions ...providers.BuildMapOption) (content []byte, err error) {
	values := buildTemplateMapMD(clusterSpec, *vs.workerNodeGroupMachineSpec)

	for _, buildOption := range buildOptions {
		buildOption(values)
	}

	bytes, err := templater.Execute(defaultClusterConfigMD, values)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func buildTemplateMapCP(clusterSpec *cluster.Spec, controlPlaneMachineSpec v1alpha1.TinkerbellMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":                  clusterSpec.ObjectMeta.Name,
		"controlPlaneEndpointIp":       clusterSpec.Spec.ControlPlaneConfiguration.Endpoint.Host,
		"controlPlaneReplicas":         clusterSpec.Spec.ControlPlaneConfiguration.Count,
		"controlPlaneSshAuthorizedKey": controlPlaneMachineSpec.Users[0].SshAuthorizedKeys,
		"controlPlaneSshUsername":      controlPlaneMachineSpec.Users[0].Name,
		"eksaSystemNamespace":          constants.EksaSystemNamespace,
		"format":                       format,
		"kubernetesVersion":            bundle.KubeDistro.Kubernetes.Tag,
		"kubeVipImage":                 bundle.VSphere.KubeVip.VersionedImage(),
		"podCidrs":                     clusterSpec.Spec.ClusterNetwork.Pods.CidrBlocks,
		"serviceCidrs":                 clusterSpec.Spec.ClusterNetwork.Services.CidrBlocks,
	}
	return values
}

func buildTemplateMapMD(clusterSpec *cluster.Spec, workerNodeGroupMachineSpec v1alpha1.TinkerbellMachineConfigSpec) map[string]interface{} {
	bundle := clusterSpec.VersionsBundle
	format := "cloud-config"

	values := map[string]interface{}{
		"clusterName":            clusterSpec.ObjectMeta.Name,
		"eksaSystemNamespace":    constants.EksaSystemNamespace,
		"format":                 format,
		"kubernetesVersion":      bundle.KubeDistro.Kubernetes.Tag,
		"workerReplicas":         clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Count,
		"workerPoolName":         "md-0",
		"workerSshAuthorizedKey": workerNodeGroupMachineSpec.Users[0].SshAuthorizedKeys,
		"workerSshUsername":      workerNodeGroupMachineSpec.Users[0].Name,
	}
	return values
}
