package tinkerbell

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// BaseControlPlane represents a CAPI Tinkerbell control plane.
type BaseControlPlane = clusterapi.ControlPlane[*tinkerbellv1.TinkerbellCluster, *tinkerbellv1.TinkerbellMachineTemplate]

// ControlPlane holds the Tinkerbell specific objects for a CAPI Tinkerbell control plane.
type ControlPlane struct {
	BaseControlPlane
	Secrets *corev1.Secret
}

// Objects returns the control plane objects associated with the Tinkerbell cluster.
func (p ControlPlane) Objects() []kubernetes.Object {
	o := p.BaseControlPlane.Objects()
	o = append(o, p.Secrets)

	return o
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
			b.ControlPlane.Secrets = obj.(*corev1.Secret)
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
		return nil, nil, errors.Wrap(err, "building tinkerbell control plane parser")
	}

	err = parser.RegisterMappings(
		yamlutil.NewMapping(constants.SecretKind, func() yamlutil.APIObject {
			return &corev1.Secret{}
		}),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "registering tinkerbell control plane mappings in parser")
	}

	builder := &controlPlaneBuilder{
		BaseBuilder:  baseBuilder,
		ControlPlane: &ControlPlane{},
	}

	return parser, builder, nil
}
