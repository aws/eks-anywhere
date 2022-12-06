package yaml

import (
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// NewControlPlaneParserAndBuilder builds a Parser and a Builder for a particular provider ControlPlane
// It registers the basic shared mappings plus another two for the provider cluster and machine template
// For ControlPlane that need to include more objects, wrap around the provider builder and implement BuildFromParsed
// Any extra mappings will need to be registered manually in the Parser.
func NewControlPlaneParserAndBuilder[C clusterapi.Object[C], M clusterapi.Object[M]](logger logr.Logger, clusterMapping yamlutil.Mapping[C], machineTemplateMapping yamlutil.Mapping[M]) (*yamlutil.Parser, *ControlPlaneBuilder[C, M], error) {
	parser := yamlutil.NewParser(logger)
	if err := RegisterControlPlaneMappings(parser); err != nil {
		return nil, nil, errors.Wrap(err, "building capi control plane parser")
	}

	err := parser.RegisterMappings(
		clusterMapping.ToAPIObjectMapping(),
		machineTemplateMapping.ToAPIObjectMapping(),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "registering provider control plane mappings")
	}

	return parser, NewControlPlaneBuilder[C, M](), nil
}

// RegisterControlPlaneMappings records the basic mappings for CAPI cluster, kubeadmcontrolplane
// and etcdadm cluster in a Parser.
func RegisterControlPlaneMappings(parser *yamlutil.Parser) error {
	err := parser.RegisterMappings(
		yamlutil.NewMapping(
			"Cluster", func() yamlutil.APIObject {
				return &clusterv1.Cluster{}
			},
		),
		yamlutil.NewMapping(
			"KubeadmControlPlane", func() yamlutil.APIObject {
				return &controlplanev1.KubeadmControlPlane{}
			},
		),
		yamlutil.NewMapping(
			"EtcdadmCluster", func() yamlutil.APIObject {
				return &etcdv1.EtcdadmCluster{}
			},
		),
	)
	if err != nil {
		return errors.Wrap(err, "registering base control plane mappings")
	}

	return nil
}

// ControlPlaneBuilder implements yamlutil.Builder
// It's a wrapper around ControlPlane to provide yaml parsing functionality.
type ControlPlaneBuilder[C clusterapi.Object[C], M clusterapi.Object[M]] struct {
	ControlPlane *clusterapi.ControlPlane[C, M]
}

// NewControlPlaneBuilder builds a ControlPlaneBuilder.
func NewControlPlaneBuilder[C clusterapi.Object[C], M clusterapi.Object[M]]() *ControlPlaneBuilder[C, M] {
	return &ControlPlaneBuilder[C, M]{
		ControlPlane: new(clusterapi.ControlPlane[C, M]),
	}
}

// BuildFromParsed reads parsed objects in ObjectLookup and sets them in the ControlPlane.
func (cp *ControlPlaneBuilder[C, M]) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	ProcessControlPlaneObjects(cp.ControlPlane, lookup)
	return nil
}

// ProcessControlPlaneObjects finds all necessary objects in the parsed objects and sets them in the ControlPlane.
func ProcessControlPlaneObjects[C clusterapi.Object[C], M clusterapi.Object[M]](cp *clusterapi.ControlPlane[C, M], lookup yamlutil.ObjectLookup) {
	ProcessCluster(cp, lookup)
	if cp.Cluster == nil {
		return
	}

	ProcessProviderCluster(cp, lookup)
	ProcessKubeadmControlPlane(cp, lookup)
	ProcessEtcdCluster(cp, lookup)
}

// ProcessCluster finds the CAPI cluster in the parsed objects and sets it in ControlPlane.
func ProcessCluster[C clusterapi.Object[C], M clusterapi.Object[M]](cp *clusterapi.ControlPlane[C, M], lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Cluster" {
			cp.Cluster = obj.(*clusterv1.Cluster)
			return
		}
	}
}

// ProcessProviderCluster finds the provider cluster in the parsed objects and sets it in ControlPlane.
func ProcessProviderCluster[C clusterapi.Object[C], M clusterapi.Object[M]](cp *clusterapi.ControlPlane[C, M], lookup yamlutil.ObjectLookup) {
	providerCluster := lookup.GetFromRef(*cp.Cluster.Spec.InfrastructureRef)
	if providerCluster == nil {
		return
	}

	cp.ProviderCluster = providerCluster.(C)
}

// ProcessKubeadmControlPlane finds the CAPI kubeadm control plane and the kubeadm control plane machine template
// in the parsed objects and sets it in ControlPlane.
func ProcessKubeadmControlPlane[C clusterapi.Object[C], M clusterapi.Object[M]](cp *clusterapi.ControlPlane[C, M], lookup yamlutil.ObjectLookup) {
	kcp := lookup.GetFromRef(*cp.Cluster.Spec.ControlPlaneRef)
	if kcp == nil {
		return
	}

	cp.KubeadmControlPlane = kcp.(*controlplanev1.KubeadmControlPlane)

	machineTemplate := lookup.GetFromRef(cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef)
	if machineTemplate == nil {
		return
	}

	cp.ControlPlaneMachineTemplate = machineTemplate.(M)
}

// ProcessEtcdCluster finds the CAPI etcdadm cluster (for unstacked clusters) in the parsed objects and sets it in ControlPlane.
func ProcessEtcdCluster[C clusterapi.Object[C], M clusterapi.Object[M]](cp *clusterapi.ControlPlane[C, M], lookup yamlutil.ObjectLookup) {
	if cp.Cluster.Spec.ManagedExternalEtcdRef == nil {
		return
	}

	etcdCluster := lookup.GetFromRef(*cp.Cluster.Spec.ManagedExternalEtcdRef)
	if etcdCluster == nil {
		return
	}

	cp.EtcdCluster = etcdCluster.(*etcdv1.EtcdadmCluster)

	etcdMachineTemplate := lookup.GetFromRef(cp.EtcdCluster.Spec.InfrastructureTemplate)
	if etcdMachineTemplate == nil {
		return
	}

	cp.EtcdMachineTemplate = etcdMachineTemplate.(M)
}
