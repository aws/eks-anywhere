package nutanix

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	nutanixv1 "github.com/nutanix-cloud-native/cluster-api-provider-nutanix/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	yamlcapi "github.com/aws/eks-anywhere/pkg/clusterapi/yaml"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// BaseControlPlane represents a CAPI Nutanix control plane.
type BaseControlPlane = clusterapi.ControlPlane[*nutanixv1.NutanixCluster, *nutanixv1.NutanixMachineTemplate]

// ControlPlane holds the Nutanix specific objects for a CAPI Nutanix control plane.
type ControlPlane struct {
	BaseControlPlane
	ConfigMaps          []*corev1.ConfigMap
	ClusterResourceSets []*addonsv1.ClusterResourceSet
}

// Objects returns the control plane objects associated with the Nutanix cluster.
func (p ControlPlane) Objects() []kubernetes.Object {
	o := p.BaseControlPlane.Objects()
	o = appendKubeObjects[*corev1.ConfigMap](o, p.ConfigMaps)
	o = appendKubeObjects[*addonsv1.ClusterResourceSet](o, p.ClusterResourceSets)

	return o
}

// ControlPlaneBuilder defines the builder for all objects in the CAPI Nutanix control plane.
type ControlPlaneBuilder struct {
	BaseBuilder  *yamlcapi.ControlPlaneBuilder[*nutanixv1.NutanixCluster, *nutanixv1.NutanixMachineTemplate]
	ControlPlane *ControlPlane
}

// BuildFromParsed implements the base yamlcapi.BuildFromParsed and processes any additional objects for the Nutanix control plane.
func (b *ControlPlaneBuilder) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	if err := b.BaseBuilder.BuildFromParsed(lookup); err != nil {
		return err
	}

	b.ControlPlane.BaseControlPlane = *b.BaseBuilder.ControlPlane
	buildObjects(b.ControlPlane, lookup)

	return nil
}

// ControlPlaneSpec builds a nutanix ControlPlane definition based on an eks-a cluster spec.
func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, spec *cluster.Spec) (*ControlPlane, error) {
	ndcs := spec.NutanixDatacenter.Spec
	machineConfigs := spec.NutanixMachineConfigs
	controlPlaneMachineSpec, etcdMachineSpec := getControlPlaneMachineSpecs(machineConfigs, &spec.Cluster.Spec.ControlPlaneConfiguration, spec.Cluster.Spec.ExternalEtcdConfiguration)
	for _, machineConfig := range machineConfigs {
		machineConfig.SetDefaults()
	}

	creds := GetCredsFromEnv()
	templateBuilder := NewNutanixTemplateBuilder(&ndcs, controlPlaneMachineSpec, etcdMachineSpec, nil, creds, time.Now)

	controlPlaneYaml, err := generateControlPlaneYAML(templateBuilder, spec)
	if err != nil {
		return nil, err
	}

	cp, err := parseControlPlaneYAML(logger, controlPlaneYaml)
	if err != nil {
		return nil, err
	}

	if err := cp.UpdateImmutableObjectNames(ctx, client, getMachineTemplate, machineTemplateEquals); err != nil {
		return nil, err
	}

	return cp, nil
}

func getControlPlaneMachineSpecs(machineConfigs map[string]*v1alpha1.NutanixMachineConfig, controlPlaneConfig *v1alpha1.ControlPlaneConfiguration, externalEtcdConfig *v1alpha1.ExternalEtcdConfiguration) (*v1alpha1.NutanixMachineConfigSpec, *v1alpha1.NutanixMachineConfigSpec) {
	var controlPlaneMachineSpec, etcdMachineSpec *v1alpha1.NutanixMachineConfigSpec
	if controlPlaneConfig.MachineGroupRef != nil && machineConfigs[controlPlaneConfig.MachineGroupRef.Name] != nil {
		controlPlaneMachineSpec = &machineConfigs[controlPlaneConfig.MachineGroupRef.Name].Spec
	}

	if externalEtcdConfig != nil && externalEtcdConfig.MachineGroupRef != nil && machineConfigs[externalEtcdConfig.MachineGroupRef.Name] != nil {
		etcdMachineSpec = &machineConfigs[externalEtcdConfig.MachineGroupRef.Name].Spec
	}

	return controlPlaneMachineSpec, etcdMachineSpec
}

func generateControlPlaneYAML(templateBuilder *TemplateBuilder, spec *cluster.Spec) ([]byte, error) {
	return templateBuilder.GenerateCAPISpecControlPlane(
		spec,
		func(values map[string]interface{}) {
			values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
			values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(spec.Cluster)
		},
	)
}

func parseControlPlaneYAML(logger logr.Logger, controlPlaneYAML []byte) (*ControlPlane, error) {
	parser, builder, err := newControlPlaneParser(logger)
	if err != nil {
		return nil, err
	}

	if err := parser.Parse(controlPlaneYAML, builder); err != nil {
		return nil, err
	}

	return builder.ControlPlane, nil
}

func newControlPlaneParser(logger logr.Logger) (*yamlutil.Parser, *ControlPlaneBuilder, error) {
	parser, baseBuilder, err := yamlcapi.NewControlPlaneParserAndBuilder(
		logger,
		yamlutil.NewMapping(
			"NutanixCluster",
			func() *nutanixv1.NutanixCluster {
				return &nutanixv1.NutanixCluster{}
			},
		),
		yamlutil.NewMapping(
			"NutanixMachineTemplate",
			func() *nutanixv1.NutanixMachineTemplate {
				return &nutanixv1.NutanixMachineTemplate{}
			},
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed building nutanix control plane parser: %w", err)
	}

	err = parser.RegisterMappings(
		yamlutil.NewMapping(
			constants.ConfigMapKind,
			func() yamlutil.APIObject {
				return &corev1.ConfigMap{}
			},
		),
		yamlutil.NewMapping(
			constants.ClusterResourceSetKind,
			func() yamlutil.APIObject {
				return &addonsv1.ClusterResourceSet{}
			},
		),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed registering nutanix control plane mappings: %w", err)
	}

	builder := &ControlPlaneBuilder{
		BaseBuilder:  baseBuilder,
		ControlPlane: &ControlPlane{},
	}

	return parser, builder, nil
}

func appendKubeObjects[V kubernetes.Object](objList []kubernetes.Object, objToAdd []V) []kubernetes.Object {
	for _, obj := range objToAdd {
		objList = append(objList, obj)
	}

	return objList
}

func buildObjects(cp *ControlPlane, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case constants.ConfigMapKind:
			cp.ConfigMaps = append(cp.ConfigMaps, obj.(*corev1.ConfigMap))
		case constants.ClusterResourceSetKind:
			cp.ClusterResourceSets = append(cp.ClusterResourceSets, obj.(*addonsv1.ClusterResourceSet))
		}
	}
}
