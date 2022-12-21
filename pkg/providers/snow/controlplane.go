package snow

import (
	"context"
	"fmt"

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
	Secret       *corev1.Secret
	CAPASIPPools CAPASIPPools
}

// Objects returns the control plane objects associated with the snow cluster.
func (c ControlPlane) Objects() []kubernetes.Object {
	o := c.BaseControlPlane.Objects()
	o = append(o, c.Secret)
	for _, p := range c.CAPASIPPools {
		o = append(o, p)
	}
	return o
}

// ControlPlaneSpec builds a snow ControlPlane definition based on an eks-a cluster spec.
func ControlPlaneSpec(ctx context.Context, logger logr.Logger, client kubernetes.Client, clusterSpec *cluster.Spec) (*ControlPlane, error) {
	capasCredentialsSecret, err := capasCredentialsSecret(clusterSpec)
	if err != nil {
		return nil, err
	}

	snowCluster := SnowCluster(clusterSpec, capasCredentialsSecret)

	cpMachineConfig := clusterSpec.SnowMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]

	capasPools := CAPASIPPools{}
	capasPools.addPools(cpMachineConfig.Spec.Network.DirectNetworkInterfaces, clusterSpec.SnowIPPools)

	cpMachineTemplate := MachineTemplate(clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster), cpMachineConfig, capasPools)

	kubeadmControlPlane, err := KubeadmControlPlane(logger, clusterSpec, cpMachineTemplate)
	if err != nil {
		return nil, err
	}

	var etcdMachineTemplate *snowv1.AWSSnowMachineTemplate
	var etcdCluster *v1beta1.EtcdadmCluster

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := clusterSpec.SnowMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		capasPools.addPools(etcdMachineConfig.Spec.Network.DirectNetworkInterfaces, clusterSpec.SnowIPPools)
		etcdMachineTemplate = MachineTemplate(clusterapi.EtcdMachineTemplateName(clusterSpec.Cluster), etcdMachineConfig, capasPools)
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
		Secret:       capasCredentialsSecret,
		CAPASIPPools: capasPools,
	}

	if err := cp.UpdateImmutableObjectNames(ctx, client, getMachineTemplate, MachineTemplateDeepDerivative); err != nil {
		return nil, errors.Wrap(err, "updating snow immutable object names")
	}

	return cp, nil
}

// credentialsSecret generates the credentials secret(s) used for provisioning a snow cluster.
// - eks-a credentials secret: user managed secret referred from snowdatacenterconfig identityRef
// - snow credentials secret: eks-a creates, updates and deletes in eksa-system namespace. this secret is fully managed by eks-a. User shall treat it as a "read-only" object.
func capasCredentialsSecret(clusterSpec *cluster.Spec) (*corev1.Secret, error) {
	if clusterSpec.SnowCredentialsSecret == nil {
		return nil, errors.New("snowCredentialsSecret in clusterSpec shall not be nil")
	}

	// we reconcile the snow credentials secret to be in sync with the eks-a credentials secret user manages.
	// notice for cli upgrade, we handle the eks-a credentials secret update in a separate step - under provider.UpdateSecrets
	// which runs before the actual cluster upgrade.
	// for controller secret, the user is responsible for making sure the eks-a credentials secret is created and up to date.
	credsB64, ok := clusterSpec.SnowCredentialsSecret.Data["credentials"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve credentials from secret [%s]", clusterSpec.SnowCredentialsSecret.GetName())
	}
	certsB64, ok := clusterSpec.SnowCredentialsSecret.Data["ca-bundle"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve ca-bundle from secret [%s]", clusterSpec.SnowCredentialsSecret.GetName())
	}

	return CAPASCredentialsSecret(clusterSpec, credsB64, certsB64), nil
}
