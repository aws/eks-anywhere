package docker

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type ControlPlane = clusterapi.ControlPlane[*dockerv1.DockerCluster, *dockerv1.DockerMachineTemplate, ProviderControlPlane]

type ProviderControlPlane struct {
	clusterapi.NoObjectsProviderControlPlane
}

func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*ControlPlane, error) {
	templateBuilder := NewDockerTemplateBuilder(time.Now)
	controlPlaneYaml, err := templateBuilder.GenerateCAPISpecControlPlane(
		spec,
		func(values map[string]interface{}) {
			values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec)
			values["etcdTemplateName"] = clusterapi.EtcdAdmMachineTemplateName(spec.Cluster)
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "generating docker control plane yaml spec")
	}

	parser, err := newControlPlaneParser(logger)
	if err != nil {
		return nil, err
	}

	cp, err := parser.Parse(controlPlaneYaml)
	if err != nil {
		return nil, errors.Wrap(err, "parsing docker control plane yaml")
	}

	if err = cp.UpdateImmutableObjectNames(ctx, client, machineTemplateComparator); err != nil {
		return nil, errors.Wrap(err, "updating docker immutable object names")
	}

	return cp, nil
}

func newControlPlaneParser(logger logr.Logger) (*yamlutil.Parser[ControlPlane], error) {
	parser, err := yamlcapi.NewControlPlaneParser[*dockerv1.DockerCluster, *dockerv1.DockerMachineTemplate, ProviderControlPlane](
		logger,
		yamlutil.NewMapping(
			"DockerCluster",
			func() *dockerv1.DockerCluster {
				return &dockerv1.DockerCluster{}
			},
		),
		yamlutil.NewMapping(
			"DockerMachineTemplate",
			func() *dockerv1.DockerMachineTemplate {
				return &dockerv1.DockerMachineTemplate{}
			},
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building docker control plane parser")
	}

	return parser, nil
}

func machineTemplateComparator(ctx context.Context, client kubernetes.Client, m *dockerv1.DockerMachineTemplate) (bool, error) {
	currentMachine := &dockerv1.DockerMachineTemplate{}
	err := client.Get(ctx, m.Name, m.Namespace, currentMachine)
	if apierrors.IsNotFound(err) {
		currentMachine = nil
	}
	if err != nil {
		return false, err
	}

	return currentMachine == nil || equality.Semantic.DeepDerivative(m.Spec, currentMachine.Spec), nil
}
