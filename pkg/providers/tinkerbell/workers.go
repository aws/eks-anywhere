package tinkerbell

import (
	"context"
	"time"

	"github.com/pkg/errors"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	capiyaml "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
	"github.com/go-logr/logr"
)

type (
	Workers        = clusterapi.Workers[*tinkerbellv1.TinkerbellMachineTemplate]
	workersBuilder = capiyaml.WorkersBuilder[*tinkerbellv1.TinkerbellMachineTemplate]
)

func WorkersSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*Workers, error) {
	wcc := spec.Cluster.Spec.WorkerNodeGroupConfigurations
	groups := make(map[string]v1alpha1.TinkerbellMachineConfigSpec, len(wcc))
	workerTemplateNames := make(map[string]string, len(wcc))
	kubeadmTemplateNames := make(map[string]string, len(wcc))
	de := hardware.NewDiskExtractor()

	for _, wc := range wcc {
		if wc.MachineGroupRef != nil && spec.TinkerbellMachineConfigs[wc.MachineGroupRef.Name] != nil {
			groups[wc.MachineGroupRef.Name] = spec.TinkerbellMachineConfigs[wc.MachineGroupRef.Name].Spec
			if err := de.Register(groups[wc.MachineGroupRef.Name].HardwareSelector); err != nil {
				return nil, err
			}
		}

		workerTemplateNames[wc.Name] = clusterapi.WorkerMachineTemplateName(spec, wc)
		kubeadmTemplateNames[wc.Name] = clusterapi.DefaultKubeadmConfigTemplateName(spec, wc)
	}

	TemplateBuilder := NewTemplateBuilder(
		&spec.TinkerbellDatacenter.Spec,
		nil,
		nil,
		de,
		groups,
		spec.TinkerbellDatacenter.Spec.TinkerbellIP,
		time.Now,
	)

	workersYaml, err := TemplateBuilder.GenerateCAPISpecWorkers(spec, workerTemplateNames, kubeadmTemplateNames)

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
