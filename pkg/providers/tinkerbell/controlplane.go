package tinkerbell

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	corev1 "k8s.io/api/core/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// BaseControlPlane represents a CAPI Tinkerbell control plane.
type BaseControlPlane = clusterapi.ControlPlane[*tinkerbellv1.TinkerbellCluster, *tinkerbellv1.TinkerbellMachineTemplate]

// ControlPlane holds the Tinkerbell specific objects for a CAPI Tinkerbell control plane.
type ControlPlane struct {
	BaseControlPlane
	Secrets []*corev1.Secret
}

// Objects returns the control plane objects associated with the Tinkerbell cluster.
func (p ControlPlane) Objects() []kubernetes.Object {
	o := p.BaseControlPlane.Objects()
	o = getSecrets(o, p.Secrets)

	return o
}

func getSecrets(o []kubernetes.Object, secrets []*corev1.Secret) []kubernetes.Object {
	for _, s := range secrets {
		o = append(o, s)
	}
	return o
}

func getControlPlaneMachineSpec(clusterSpec *cluster.Spec, diskExtractor *hardware.DiskExtractor) (*anywherev1.TinkerbellMachineConfigSpec, error) {
	var controlPlaneMachineSpec *anywherev1.TinkerbellMachineConfigSpec
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
		err := diskExtractor.Register(controlPlaneMachineSpec.HardwareSelector)
		if err != nil {
			return nil, err
		}
	}

	return controlPlaneMachineSpec, nil
}

func getWorkerNodeGroupMachineSpec(clusterSpec *cluster.Spec, diskExtractor *hardware.DiskExtractor) (map[string]anywherev1.TinkerbellMachineConfigSpec, error) {
	var workerNodeGroupMachineSpec *anywherev1.TinkerbellMachineConfigSpec
	workerNodeGroupMachineSpecs := make(map[string]anywherev1.TinkerbellMachineConfigSpec, len(clusterSpec.TinkerbellMachineConfigs))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if wnConfig.MachineGroupRef != nil && clusterSpec.TinkerbellMachineConfigs[wnConfig.MachineGroupRef.Name] != nil {
			workerNodeGroupMachineSpec = &clusterSpec.TinkerbellMachineConfigs[wnConfig.MachineGroupRef.Name].Spec
			workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = *workerNodeGroupMachineSpec
			err := diskExtractor.Register(workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name].HardwareSelector)
			if err != nil {
				return nil, err
			}
		}
	}

	return workerNodeGroupMachineSpecs, nil
}

func getEtcdMachineSpec(clusterSpec *cluster.Spec, diskExtractor *hardware.DiskExtractor) (*anywherev1.TinkerbellMachineConfigSpec, error) {
	var etcdMachineSpec *anywherev1.TinkerbellMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] != nil {
			etcdMachineSpec = &clusterSpec.TinkerbellMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
			err := diskExtractor.Register(etcdMachineSpec.HardwareSelector)
			if err != nil {
				return nil, err
			}
		}
	}

	return etcdMachineSpec, nil
}

func generateTemplateBuilder(clusterSpec *cluster.Spec) (providers.TemplateBuilder, error) {
	diskExtractor := hardware.NewDiskExtractor()

	controlPlaneMachineSpec, err := getControlPlaneMachineSpec(clusterSpec, diskExtractor)
	if err != nil {
		return nil, errors.Wrap(err, "generating control plane machine spec")
	}

	workerNodeGroupMachineSpecs, err := getWorkerNodeGroupMachineSpec(clusterSpec, diskExtractor)
	if err != nil {
		return nil, errors.Wrap(err, "generating worker node group machine specs")
	}

	etcdMachineSpec, err := getEtcdMachineSpec(clusterSpec, diskExtractor)
	if err != nil {
		return nil, errors.Wrap(err, "generating etcd machine spec")
	}

	templateBuilder := NewTemplateBuilder(&clusterSpec.TinkerbellDatacenter.Spec,
		controlPlaneMachineSpec,
		etcdMachineSpec,
		diskExtractor,
		workerNodeGroupMachineSpecs,
		clusterSpec.TinkerbellDatacenter.Spec.TinkerbellIP,
		time.Now)
	return templateBuilder, nil
}

// ControlPlaneSpec builds a Tinkerbell ControlPlane definition based on an eks-a cluster spec.
func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, clusterSpec *cluster.Spec) (*ControlPlane, error) {
	templateBuilder, err := generateTemplateBuilder(clusterSpec)
	if err != nil {
		return nil, errors.Wrap(err, "generating tinkerbell template builder")
	}

	controlPlaneYaml, err := templateBuilder.GenerateCAPISpecControlPlane(
		clusterSpec,
		func(values map[string]interface{}) {
			values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster)
			values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(clusterSpec.Cluster)
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "generating tinkerbell control plane yaml spec")
	}

	parser, builder, err := newControlPlaneParser(logger)
	if err != nil {
		return nil, err
	}

	err = parser.Parse(controlPlaneYaml, builder)
	if err != nil {
		return nil, errors.Wrap(err, "parsing tinkerbell control plane yaml")
	}

	cp := builder.ControlPlane
	if err = cp.UpdateImmutableObjectNames(ctx, client, GetMachineTemplate, machineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating tinkerbell immutable object names")
	}

	return cp, nil
}

// ControlPlaneBuilder defines the builder for all objects in the CAPI Tinkerbell control plane.
type controlPlaneBuilder struct {
	BaseBuilder  *yamlcapi.ControlPlaneBuilder[*tinkerbellv1.TinkerbellCluster, *tinkerbellv1.TinkerbellMachineTemplate]
	ControlPlane *ControlPlane
}

// BuildFromParsed implements the base yamlcapi.BuildFromParsed and processes any additional objects (secrets) for the Tinkerbell control plane.
func (b *controlPlaneBuilder) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	if err := b.BaseBuilder.BuildFromParsed(lookup); err != nil {
		return err
	}

	b.ControlPlane.BaseControlPlane = *b.BaseBuilder.ControlPlane
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == constants.SecretKind {
			b.ControlPlane.Secrets = append(b.ControlPlane.Secrets, obj.(*corev1.Secret))
		}
	}

	return nil
}

func newControlPlaneParser(logger logr.Logger) (*yamlutil.Parser, *controlPlaneBuilder, error) {
	parser, baseBuilder, err := yamlcapi.NewControlPlaneParserAndBuilder(
		logger,
		yamlutil.NewMapping(
			"TinkerbellCluster",
			func() *tinkerbellv1.TinkerbellCluster {
				return &tinkerbellv1.TinkerbellCluster{}
			},
		),
		yamlutil.NewMapping(
			"TinkerbellMachineTemplate",
			func() *tinkerbellv1.TinkerbellMachineTemplate {
				return &tinkerbellv1.TinkerbellMachineTemplate{}
			},
		),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error building tinkerbell control plane parser")
	}

	builder := &controlPlaneBuilder{
		BaseBuilder:  baseBuilder,
		ControlPlane: &ControlPlane{},
	}

	return parser, builder, nil
}
