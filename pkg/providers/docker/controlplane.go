package docker

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// ControlPlane represents a CAPI Docker control plane.
type ControlPlane = clusterapi.ControlPlane[*dockerv1.DockerCluster, *dockerv1.DockerMachineTemplate]

type controlPlaneBuilder = yamlcapi.ControlPlaneBuilder[*dockerv1.DockerCluster, *dockerv1.DockerMachineTemplate]

// ControlPlaneSpec builds a docker ControlPlane definition based on an eks-a cluster spec.
func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*ControlPlane, error) {
	templateBuilder := NewDockerTemplateBuilder(time.Now)

	controlPlaneYaml, err := templateBuilder.GenerateCAPISpecControlPlane(
		spec,
		func(values map[string]interface{}) {
			values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
			values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(spec.Cluster)
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "generating docker control plane yaml spec")
	}

	parser, builder, err := newControlPlaneParser(logger)
	if err != nil {
		return nil, err
	}

	err = parser.Parse(controlPlaneYaml, builder)
	if err != nil {
		return nil, errors.Wrap(err, "parsing docker control plane yaml")
	}

	cp := builder.ControlPlane
	if err = cp.UpdateImmutableObjectNames(ctx, client, GetMachineTemplate, MachineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating docker immutable object names")
	}

	return cp, nil
}

func newControlPlaneParser(logger logr.Logger) (*yamlutil.Parser, *controlPlaneBuilder, error) {
	parser, builder, err := yamlcapi.NewControlPlaneParserAndBuilder(
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
		return nil, nil, errors.Wrap(err, "building docker control plane parser")
	}

	return parser, builder, nil
}
