package vsphere

import (
	corev1 "k8s.io/api/core/v1"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
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
	Secrets            []*corev1.Secret
	ConfigMaps         []*corev1.ConfigMap
	ClusterResourceSet *clusterapi.ClusterResourceSet
}

// Objects returns the control plane objects associated with the VSphere cluster.
func (p ControlPlane) Objects() []kubernetes.Object {
	o := p.BaseControlPlane.Objects()
	o = getSecrets(o, p.Secrets)
	o = getConfigMaps(o, p.ConfigMaps)
	// TODO: Get ClusterResourceSet

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
	processSecrets(b.ControlPlane, lookup)
	processConfigMaps(b.ControlPlane, lookup)
	// TODO: Process ClusterResourceSet

	return nil
}

func processSecrets(c *ControlPlane, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == constants.SecretKind {
			c.Secrets = append(c.Secrets, obj.(*corev1.Secret))
		}
	}
}

func processConfigMaps(c *ControlPlane, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == constants.ConfigMapKind {
			c.ConfigMaps = append(c.ConfigMaps, obj.(*corev1.ConfigMap))
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
