package vsphere

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	capiyaml "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type (
	// Workers represents the vSphere specific CAPI spec for worker nodes.
	Workers        = clusterapi.Workers[*vspherev1.VSphereMachineTemplate]
	workersBuilder = capiyaml.WorkersBuilder[*vspherev1.VSphereMachineTemplate]
)

// WorkersSpec generates a vSphere specific CAPI spec for an eks-a cluster worker nodes.
// It talks to the cluster with a client to detect changes in immutable objects and generates new
// names for them.
func WorkersSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*Workers, error) {
	// TODO(g-gaston): refactor template builder so it doesn't behave differently for controller and CLI
	// TODO(g-gaston): do we need time.Now if the names are not dependent on a timestamp anymore?
	templateBuilder := NewVsphereTemplateBuilder(time.Now)
	workersYaml, err := templateBuilder.CAPIWorkersSpecWithInitialNames(spec)
	if err != nil {
		return nil, err
	}

	parser, builder, err := newWorkersParserAndBuilder(logger)
	if err != nil {
		return nil, err
	}

	if err = parser.Parse(workersYaml, builder); err != nil {
		return nil, errors.Wrap(err, "parsing vSphere CAPI workers yaml")
	}

	workers := builder.Workers
	if err = workers.UpdateImmutableObjectNames(ctx, client, getMachineTemplate, machineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating vSphere worker immutable object names")
	}

	return workers, nil
}

func newWorkersParserAndBuilder(logger logr.Logger) (*yamlutil.Parser, *workersBuilder, error) {
	parser, builder, err := capiyaml.NewWorkersParserAndBuilder(
		logger,
		machineTemplateMapping(),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "building vSphere workers parser and builder")
	}

	return parser, builder, nil
}

func machineTemplateMapping() yamlutil.Mapping[*vspherev1.VSphereMachineTemplate] {
	return yamlutil.NewMapping(
		"VSphereMachineTemplate",
		func() *vspherev1.VSphereMachineTemplate {
			return &vspherev1.VSphereMachineTemplate{}
		},
	)
}
