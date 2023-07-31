package cloudstack

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	capiyaml "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type (
	// Workers represents the cloudstack specific CAPI spec for worker nodes.
	Workers        = clusterapi.Workers[*cloudstackv1.CloudStackMachineTemplate]
	workersBuilder = capiyaml.WorkersBuilder[*cloudstackv1.CloudStackMachineTemplate]
)

// WorkersSpec generates a cloudstack specific CAPI spec for an eks-a cluster worker nodes.
// It talks to the cluster with a client to detect changes in immutable objects and generates new
// names for them.
func WorkersSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*Workers, error) {
	templateBuilder := NewTemplateBuilder(time.Now)
	machineTemplateNames, kubeadmConfigTemplateNames := clusterapi.InitialTemplateNamesForWorkers(spec)
	workersYaml, err := templateBuilder.GenerateCAPISpecWorkers(spec, machineTemplateNames, kubeadmConfigTemplateNames)
	if err != nil {
		return nil, err
	}

	parser, builder, err := newWorkersParserAndBuilder(logger)
	if err != nil {
		return nil, err
	}

	if err = parser.Parse(workersYaml, builder); err != nil {
		return nil, errors.Wrap(err, "parsing cloudstack CAPI workers yaml")
	}

	workers := builder.Workers
	if err = workers.UpdateImmutableObjectNames(ctx, client, GetMachineTemplate, machineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating cloudstack worker immutable object names")
	}

	return workers, nil
}

func newWorkersParserAndBuilder(logger logr.Logger) (*yamlutil.Parser, *workersBuilder, error) {
	parser, builder, err := capiyaml.NewWorkersParserAndBuilder(
		logger,
		machineTemplateMapping(),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "building cloudstack workers parser and builder")
	}

	return parser, builder, nil
}

func machineTemplateMapping() yamlutil.Mapping[*cloudstackv1.CloudStackMachineTemplate] {
	return yamlutil.NewMapping(
		"CloudStackMachineTemplate",
		func() *cloudstackv1.CloudStackMachineTemplate {
			return &cloudstackv1.CloudStackMachineTemplate{}
		},
	)
}
