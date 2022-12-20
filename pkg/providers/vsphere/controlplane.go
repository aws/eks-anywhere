package vsphere

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// BaseControlPlane represents a CAPI VSphere control plane.
type BaseControlPlane = clusterapi.ControlPlane[*vspherev1.VSphereCluster, *vspherev1.VSphereMachineTemplate]

// ControlPlane holds the VSphere specific objects for a CAPI VSphere control plane.
type ControlPlane struct {
	BaseControlPlane
	Secrets             []*corev1.Secret
	ConfigMaps          []*corev1.ConfigMap
	ClusterResourceSets []*addonsv1.ClusterResourceSet
}

// Objects returns the control plane objects associated with the VSphere cluster.
func (p ControlPlane) Objects() []kubernetes.Object {
	o := p.BaseControlPlane.Objects()
	o = getSecrets(o, p.Secrets)
	o = getConfigMaps(o, p.ConfigMaps)
	o = getClusterResourceSets(o, p.ClusterResourceSets)

	return o
}

// ControlPlaneBuilder defines the builder for all objects in the CAPI VSphere control plane.
type ControlPlaneBuilder struct {
	BaseBuilder  *yamlcapi.ControlPlaneBuilder[*vspherev1.VSphereCluster, *vspherev1.VSphereMachineTemplate]
	ControlPlane *ControlPlane
}

// BuildFromParsed implements the base yamlcapi.BuildFromParsed and processes any additional objects for the VSphere control plane.
func (b *ControlPlaneBuilder) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	if err := b.BaseBuilder.BuildFromParsed(lookup); err != nil {
		return err
	}

	b.ControlPlane.BaseControlPlane = *b.BaseBuilder.ControlPlane
	processObjects(b.ControlPlane, lookup)

	return nil
}

// ControlPlaneSpec builds a vsphere ControlPlane definition based on an eks-a cluster spec.
func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*ControlPlane, error) {
	templateBuilder := NewVsphereTemplateBuilder(time.Now)

	controlPlaneYaml, err := templateBuilder.GenerateCAPISpecControlPlane(
		spec,
		func(values map[string]interface{}) {
			values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
			values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(spec.Cluster)
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "generating vsphere control plane yaml spec")
	}

	parser, builder, err := newControlPlaneParser(logger)
	if err != nil {
		return nil, err
	}

	err = parser.Parse(controlPlaneYaml, builder)
	if err != nil {
		return nil, errors.Wrap(err, "parsing vsphere control plane yaml")
	}

	cp := builder.ControlPlane

	if err = cp.UpdateImmutableObjectNames(ctx, client, getMachineTemplate, machineTemplateEqual); err != nil {
		return nil, errors.Wrap(err, "updating vsphere immutable object names")
	}

	return cp, nil
}

func newControlPlaneParser(logger logr.Logger) (*yamlutil.Parser, *ControlPlaneBuilder, error) {
	parser, baseBuilder, err := yamlcapi.NewControlPlaneParserAndBuilder(
		logger,
		yamlutil.NewMapping(
			"VSphereCluster",
			func() *vspherev1.VSphereCluster {
				return &vspherev1.VSphereCluster{}
			},
		),
		yamlutil.NewMapping(
			"VSphereMachineTemplate",
			func() *vspherev1.VSphereMachineTemplate {
				return &vspherev1.VSphereMachineTemplate{}
			},
		),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "building vsphere control plane parser")
	}

	err = parser.RegisterMappings(
		yamlutil.NewMapping(constants.SecretKind, func() yamlutil.APIObject {
			return &corev1.Secret{}
		}),
		yamlutil.NewMapping(constants.ConfigMapKind, func() yamlutil.APIObject {
			return &corev1.ConfigMap{}
		}),
		yamlutil.NewMapping(constants.ClusterResourceSetKind, func() yamlutil.APIObject {
			return &addonsv1.ClusterResourceSet{}
		}),
	)

	if err != nil {
		return nil, nil, errors.Wrap(err, "registering vsphere control plane mappings in parser")
	}

	builder := &ControlPlaneBuilder{
		BaseBuilder:  baseBuilder,
		ControlPlane: &ControlPlane{},
	}

	return parser, builder, nil
}

func processObjects(c *ControlPlane, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case constants.SecretKind:
			c.Secrets = append(c.Secrets, obj.(*corev1.Secret))
		case constants.ConfigMapKind:
			c.ConfigMaps = append(c.ConfigMaps, obj.(*corev1.ConfigMap))
		case constants.ClusterResourceSetKind:
			c.ClusterResourceSets = append(c.ClusterResourceSets, obj.(*addonsv1.ClusterResourceSet))
		}
	}
}

func getSecrets(o []kubernetes.Object, secrets []*corev1.Secret) []kubernetes.Object {
	for _, s := range secrets {
		o = append(o, s)
	}
	return o
}

func getConfigMaps(o []kubernetes.Object, configMaps []*corev1.ConfigMap) []kubernetes.Object {
	for _, m := range configMaps {
		o = append(o, m)
	}
	return o
}

func getClusterResourceSets(o []kubernetes.Object, clusterResourceSets []*addonsv1.ClusterResourceSet) []kubernetes.Object {
	for _, s := range clusterResourceSets {
		o = append(o, s)
	}
	return o
}
