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
	capiyaml "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type (
	// Workers represents the docker specific CAPI spec for worker nodes.
	Workers        = clusterapi.Workers[*dockerv1.DockerMachineTemplate]
	workersBuilder = capiyaml.WorkersBuilder[*dockerv1.DockerMachineTemplate]
)

// WorkersSpec generates a Docker specific CAPI spec for an eks-a cluster worker nodes.
// It talks to the cluster with a client to detect changes in immutable objects and generates new
// names for them.
func WorkersSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*Workers, error) {
	templateBuilder := NewDockerTemplateBuilder(time.Now)
	workersYaml, err := templateBuilder.CAPIWorkersSpecWithInitialNames(spec)
	if err != nil {
		return nil, err
	}

	parser, builder, err := newWorkersParserAndBuilder(logger)
	if err != nil {
		return nil, err
	}

	if err = parser.Parse(workersYaml, builder); err != nil {
		return nil, errors.Wrap(err, "parsing docker CAPI workers yaml")
	}

	workers := builder.Workers
	if err = workers.UpdateImmutableObjectNames(ctx, client, GetMachineTemplate, MachineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating docker worker immutable object names")
	}

	return workers, nil
}

func newWorkersParserAndBuilder(logger logr.Logger) (*yamlutil.Parser, *workersBuilder, error) {
	parser, builder, err := capiyaml.NewWorkersParserAndBuilder(
		logger,
		machineTemplateMapping(),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "building docker workers parser and builder")
	}

	return parser, builder, nil
}

func machineTemplateMapping() yamlutil.Mapping[*dockerv1.DockerMachineTemplate] {
	return yamlutil.NewMapping(
		"DockerMachineTemplate",
		func() *dockerv1.DockerMachineTemplate {
			return &dockerv1.DockerMachineTemplate{}
		},
	)
}
