package docker

import (
	"time"

	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type ControlPlane = clusterapi.ControlPlane[ProviderControlPlane]

type ProviderControlPlane struct {
	cluster             *dockerv1.DockerCluster
	machineTemplate     *dockerv1.DockerMachineTemplate
	etcdMachineTemplate *dockerv1.DockerMachineTemplate
}

func (cp ProviderControlPlane) Objects() []kubernetes.Object {
	return []kubernetes.Object{cp.cluster, cp.machineTemplate, cp.etcdMachineTemplate}
}

func ControlPlaneSpec(logger logr.Logger, spec *cluster.Spec) (*ControlPlane, error) {
	templateBuilder := NewDockerTemplateBuilder(time.Now)
	controlPlaneYaml, err := templateBuilder.GenerateCAPISpecControlPlane(spec)
	if err != nil {
		return nil, err
	}

	parser, err := newControlPlaneParser(logger)
	if err != nil {
		return nil, err
	}

	return parser.Parse(controlPlaneYaml)
}

func newControlPlaneParser(logger logr.Logger) (*yamlutil.Parser[ControlPlane], error) {
	parser := yamlutil.NewParser[ControlPlane](logger)

	if err := clusterapi.RegisterControlPlaneMappings(parser); err != nil {
		return nil, err
	}

	err := parser.RegisterMappings(map[string]yamlutil.APIObjectGenerator{
		"DockerCluster": func() yamlutil.APIObject {
			return &dockerv1.DockerCluster{}
		},
		"DockerMachineTemplate": func() yamlutil.APIObject {
			return &dockerv1.DockerMachineTemplate{}
		},
	})
	if err != nil {
		return nil, err
	}

	parser.RegisterProcessors(
		// Order is important, register CAPICluster before anything else
		clusterapi.ProcessCluster[ProviderControlPlane],
		clusterapi.ProcessKubeadmControlPlane[ProviderControlPlane],
		clusterapi.ProcessEtcdCluster[ProviderControlPlane],
		ProcessCluster,
		ProcessEtcdMachineTemplate,
	)

	return parser, nil
}

// I suspect this is going to be very similar in all providers
// Maybe it's worth moving this logic to the clusterapi package and
// Parametrize the CP with the provider cluster and machinetemplate types
func ProcessCluster(cp *ControlPlane, lookup yamlutil.ObjectLookup) {
	if cp.Cluster == nil || cp.KubeadmControlPlane == nil {
		return
	}

	dockerCluster := lookup.GetFromRef(*cp.Cluster.Spec.InfrastructureRef)
	if dockerCluster == nil {
		return
	}

	cp.Provider.cluster = dockerCluster.(*dockerv1.DockerCluster)

	machineTemplate := lookup.GetFromRef(cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef)
	if machineTemplate == nil {
		return
	}

	cp.Provider.machineTemplate = machineTemplate.(*dockerv1.DockerMachineTemplate)
}

func ProcessEtcdMachineTemplate(cp *ControlPlane, lookup yamlutil.ObjectLookup) {
	if cp.EtcdCluster == nil {
		return
	}

	etcdMachineTemplate := lookup.GetFromRef(cp.EtcdCluster.Spec.InfrastructureTemplate)
	if etcdMachineTemplate == nil {
		return
	}

	cp.Provider.etcdMachineTemplate = etcdMachineTemplate.(*dockerv1.DockerMachineTemplate)
}
