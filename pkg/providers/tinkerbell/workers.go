package tinkerbell

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	capiyaml "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type (
	// Workers represents the Tinkerbell specific CAPI spec for worker nodes.
	Workers        = clusterapi.Workers[*tinkerbellv1.TinkerbellMachineTemplate]
	workersBuilder = capiyaml.WorkersBuilder[*tinkerbellv1.TinkerbellMachineTemplate]
)

// WorkersSpec generates a Tinkerbell specific CAPI spec for an eks-a cluster worker nodes.
// It talks to the cluster with a client to detect changes in immutable objects and generates new
// names for them.
func WorkersSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*Workers, error) {
	templateBuilder, err := generateTemplateBuilder(spec)
	if err != nil {
		return nil, errors.Wrap(err, "generating tinkerbell template builder")
	}

	workerTemplateNames, kubeadmTemplateNames := clusterapi.InitialTemplateNamesForWorkers(spec)
	workersYaml, err := templateBuilder.GenerateCAPISpecWorkers(spec, workerTemplateNames, kubeadmTemplateNames)
	if err != nil {
		return nil, errors.Wrap(err, "generating tinkerbell workers yaml spec")
	}

	parser, builder, err := newWorkersParserAndBuilder(logger)
	if err != nil {
		return nil, err
	}

	if err = parser.Parse(workersYaml, builder); err != nil {
		return nil, errors.Wrap(err, "parsing Tinkerbell CAPI workers yaml")
	}

	workers := builder.Workers
	if err = workers.UpdateImmutableObjectNames(ctx, client, GetMachineTemplate, machineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating Tinkerbell worker immutable object names")
	}

	return workers, nil
}

func newWorkersParserAndBuilder(logger logr.Logger) (*yamlutil.Parser, *workersBuilder, error) {
	parser, builder, err := capiyaml.NewWorkersParserAndBuilder(
		logger,
		machineTemplateMapping(),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "building Tinkerbell workers parser and builder")
	}

	return parser, builder, nil
}

func machineTemplateMapping() yamlutil.Mapping[*tinkerbellv1.TinkerbellMachineTemplate] {
	return yamlutil.NewMapping(
		"TinkerbellMachineTemplate",
		func() *tinkerbellv1.TinkerbellMachineTemplate {
			return &tinkerbellv1.TinkerbellMachineTemplate{}
		},
	)
}
