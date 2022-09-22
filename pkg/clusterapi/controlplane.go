package clusterapi

import (
	"context"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// ProviderControlPlane represents the provider specific part of a CAPI control plane definition
type ProviderControlPlane interface {
	// Objects returns all provider specific API objects
	Objects() []kubernetes.Object
}

// NoObjectsProviderControlPlane is a helper for providers that don't require extra objects in the control plane
type NoObjectsProviderControlPlane struct{}

func (NoObjectsProviderControlPlane) Objects() []kubernetes.Object {
	return nil
}

// Object represents a kubernetes API object
type Object interface {
	kubernetes.Object
}

// ControlPlane represents the spec for a CAPI control plane for an specific provider
type ControlPlane[C, M Object, P ProviderControlPlane] struct {
	Cluster                     *clusterv1.Cluster
	ProviderCluster             C
	KubeadmControlPlane         *controlplanev1.KubeadmControlPlane
	ControlPlaneMachineTemplate M
	EtcdMachineTemplate         M
	EtcdCluster                 *etcdv1.EtcdadmCluster

	// Provider holds the provider specific components for the control plane
	Provider P
}

// Objects returns all API objects that form a concrete control plane for an specific provider
func (cp *ControlPlane[C, M, P]) Objects() []kubernetes.Object {
	objs := cp.Provider.Objects()
	objs = append(objs, cp.Cluster, cp.KubeadmControlPlane)
	if cp.EtcdCluster != nil {
		objs = append(objs, cp.EtcdCluster)
	}

	return objs
}

// UpdateImmutableObjectNames checks if any control plane immutable objects have changed y comparing the new definition
// with the current state of the cluster. If they had, it generates a new name for them by increasing a monotonic number
// at the end of the name
// This is applied to all provider machine templates
// machineTemplateComparator receives the new definition of a machine template with the name of the current one
// if should retrieve such machine template from the cluster and return true only if only both machine templates are identical
// Most of the time, this only requires comparing the Spec field, but that can change from provider to provider
func (cp *ControlPlane[C, M, P]) UpdateImmutableObjectNames(
	ctx context.Context,
	client kubernetes.Client,
	machineTemplateComparator func(context.Context, kubernetes.Client, M) (bool, error),
) error {
	currentKCP := &controlplanev1.KubeadmControlPlane{}
	err := client.Get(ctx, cp.KubeadmControlPlane.Name, cp.KubeadmControlPlane.Namespace, currentKCP)
	if apierrors.IsNotFound(err) {
		// KubeadmControlPlane doesn't exist, this is a new cluster so machine templates should use their default name
		return nil
	}
	if err != nil {
		return err
	}

	cp.ControlPlaneMachineTemplate.SetName(currentKCP.Spec.MachineTemplate.InfrastructureRef.Name)
	if err = ensureUniqueObjectName(ctx, client, machineTemplateComparator, cp.ControlPlaneMachineTemplate); err != nil {
		return err
	}

	if cp.EtcdCluster == nil {
		return nil
	}

	currentEtcdCluster := &etcdv1.EtcdadmCluster{}
	err = client.Get(ctx, cp.EtcdCluster.Name, cp.EtcdCluster.Namespace, currentEtcdCluster)
	if apierrors.IsNotFound(err) {
		// EtcdadmCluster doesn't exist, this is a new cluster so machine templates should use their default name
		return nil
	}
	if err != nil {
		return err
	}

	cp.EtcdMachineTemplate.SetName(currentEtcdCluster.Spec.InfrastructureTemplate.Name)
	if err = ensureUniqueObjectName(ctx, client, machineTemplateComparator, cp.EtcdMachineTemplate); err != nil {
		return err
	}

	return nil
}

func ensureUniqueObjectName[M Object](ctx context.Context,
	client kubernetes.Client,
	machineTemplateComparator func(context.Context, kubernetes.Client, M) (bool, error),
	machineTemplate M,
) error {
	equal, err := machineTemplateComparator(ctx, client, machineTemplate)
	if err != nil {
		return err
	}

	if !equal {
		newName, err := IncrementName(machineTemplate.GetName())
		if err != nil {
			return err
		}

		machineTemplate.SetName(newName)
	}

	return nil
}
