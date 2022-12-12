package snow

import (
	"context"

	"github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

// BaseControlPlane represents a CAPI Snow control plane.
type BaseControlPlane = clusterapi.ControlPlane[*snowv1.AWSSnowCluster, *snowv1.AWSSnowMachineTemplate]

// ControlPlane holds the Snow specific objects for a CAPI snow control plane.
type ControlPlane struct {
	BaseControlPlane
	Secret *corev1.Secret
}

// Objects returns the control plane objects associated with the snow cluster.
func (c ControlPlane) Objects() []kubernetes.Object {
	o := c.BaseControlPlane.Objects()
	o = append(o, c.Secret)
	return o
}

// ControlPlaneSpec builds a snow ControlPlane definition based on an eks-a cluster spec.
func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, clusterSpec *cluster.Spec) (*ControlPlane, error) {
	capasCredentialsSecret, err := capasCredentialsSecret(clusterSpec)
	if err != nil {
		return nil, err
	}

	snowCluster := SnowCluster(clusterSpec, capasCredentialsSecret)

	cpMachineTemplate := SnowMachineTemplate(clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster), clusterSpec.SnowMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])

	kubeadmControlPlane, err := KubeadmControlPlane(logger, clusterSpec, cpMachineTemplate)
	if err != nil {
		return nil, err
	}

	var etcdMachineTemplate *snowv1.AWSSnowMachineTemplate
	var etcdCluster *v1beta1.EtcdadmCluster

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineTemplate = SnowMachineTemplate(clusterapi.EtcdMachineTemplateName(clusterSpec.Cluster), clusterSpec.SnowMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name])
		etcdCluster = EtcdadmCluster(logger, clusterSpec, etcdMachineTemplate)
	}

	capiCluster := CAPICluster(clusterSpec, snowCluster, kubeadmControlPlane, etcdCluster)

	cp := &ControlPlane{
		BaseControlPlane: BaseControlPlane{
			Cluster:                     capiCluster,
			ProviderCluster:             snowCluster,
			KubeadmControlPlane:         kubeadmControlPlane,
			ControlPlaneMachineTemplate: cpMachineTemplate,
			EtcdCluster:                 etcdCluster,
			EtcdMachineTemplate:         etcdMachineTemplate,
		},
		Secret: capasCredentialsSecret,
	}

	if err := cp.UpdateImmutableObjectNames(ctx, client, getMachineTemplate, MachineTemplateDeepDerivative); err != nil {
		return nil, errors.Wrap(err, "updating snow immutable object names")
	}

	return cp, nil
}
