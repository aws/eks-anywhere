package yaml

import (
	"github.com/go-logr/logr"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// ProcessCluster finds the CAPI cluster in the parsed objects and sets it in ControlPlane
func ProcessCluster[C, M clusterapi.Object, P clusterapi.ProviderControlPlane](cp *clusterapi.ControlPlane[C, M, P], lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Cluster" {
			cp.Cluster = obj.(*clusterv1.Cluster)
			return
		}
	}
}

// ProcessCluster finds the provider cluster and the kubeadm control plane machine template in the parsed objects
// and sets it in ControlPlane
// Both Cluster and KubeadmControlPlane have to be processed before this
func ProcessProviderCluster[C, M clusterapi.Object, P clusterapi.ProviderControlPlane](cp *clusterapi.ControlPlane[C, M, P], lookup yamlutil.ObjectLookup) {
	if cp.Cluster == nil || cp.KubeadmControlPlane == nil {
		return
	}

	providerCluster := lookup.GetFromRef(*cp.Cluster.Spec.InfrastructureRef)
	if providerCluster == nil {
		return
	}

	cp.ProviderCluster = providerCluster.(C)

	machineTemplate := lookup.GetFromRef(cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef)
	if machineTemplate == nil {
		return
	}

	cp.ControlPlaneMachineTemplate = machineTemplate.(M)
}

// ProcessKubeadmControlPlane finds the CAPI kubeadm control plane in the parsed objects and sets it in ControlPlane
func ProcessKubeadmControlPlane[C, M clusterapi.Object, P clusterapi.ProviderControlPlane](cp *clusterapi.ControlPlane[C, M, P], lookup yamlutil.ObjectLookup) {
	if cp.Cluster == nil {
		return
	}

	kcp := lookup.GetFromRef(*cp.Cluster.Spec.ControlPlaneRef)
	if kcp == nil {
		return
	}

	cp.KubeadmControlPlane = kcp.(*controlplanev1.KubeadmControlPlane)
}

// ProcessEtcdCluster finds the CAPI etcdadm cluster (for unstacked clusters) in the parsed objects and sets it in ControlPlane
func ProcessEtcdCluster[C, M clusterapi.Object, P clusterapi.ProviderControlPlane](cp *clusterapi.ControlPlane[C, M, P], lookup yamlutil.ObjectLookup) {
	if cp.Cluster == nil || cp.Cluster.Spec.ManagedExternalEtcdRef == nil {
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

// RegisterControlPlaneMappings records the basic mappings for CAPI cluster, kubeadmcontrolplane
// and etcdadm cluster in a Parser
func RegisterControlPlaneMappings[T any](parser *yamlutil.Parser[T]) error {
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
		return errors.Wrap(err, "registering base control plan mappings")
	}

	return nil
}

// NewControlPlaneParser builds a Parser for a particular provider ControlPlane
// It registers the basic shared mappings plus the provider cluster and machine template ones
// Any extra mappings or processors for objects in the ProviderControlPlane (P) will need to be
// registered manually
func NewControlPlaneParser[C, M clusterapi.Object, P clusterapi.ProviderControlPlane](logger logr.Logger, clusterMapping yamlutil.Mapping[C], machineConfigMapping yamlutil.Mapping[M]) (*yamlutil.Parser[clusterapi.ControlPlane[C, M, P]], error) {
	parser := yamlutil.NewParser[clusterapi.ControlPlane[C, M, P]](logger)
	if err := RegisterControlPlaneMappings(parser); err != nil {
		return nil, errors.Wrap(err, "building capi control plane parser")
	}

	err := parser.RegisterMappings(
		clusterMapping.ToGenericMapping(),
		machineConfigMapping.ToGenericMapping(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "registering provider mappings")
	}

	parser.RegisterProcessors(
		// Order is important, register CAPICluster before anything else
		ProcessCluster[C, M, P],
		ProcessKubeadmControlPlane[C, M, P],
		ProcessEtcdCluster[C, M, P],
		ProcessProviderCluster[C, M, P],
	)

	return parser, nil
}
