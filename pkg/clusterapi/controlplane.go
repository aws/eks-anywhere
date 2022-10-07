package clusterapi

import (
	"context"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// Object represents a kubernetes API object
type Object interface {
	kubernetes.Object
}

// ObjectComparator returns true only if only both kubernetes Object's are identical
// Most of the time, this only requires comparing the Spec field, but that can variate
// from object to object
type ObjectComparator[O Object] func(current, new O) bool

// ObjectRetriever gets a kubernetes API object using the provided client
// If the object doesn't exist, it returns a NotFound error
type ObjectRetriever[O Object] func(ctx context.Context, client kubernetes.Client, name, namespace string) (O, error)

// ControlPlane represents the provider-specific spec for a CAPI control plane using the kubeadm CP provider
type ControlPlane[C, M Object] struct {
	Cluster *clusterv1.Cluster

	// ProviderCluster is the provider-specific resource that holds the details
	// for provisioning the infrastructure, referenced in Cluster.Spec.InfrastructureRef
	ProviderCluster C

	KubeadmControlPlane *controlplanev1.KubeadmControlPlane

	// ControlPlaneMachineTemplate is the provider-specific machine template referenced
	// in KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef
	ControlPlaneMachineTemplate M

	EtcdCluster *etcdv1.EtcdadmCluster

	// EtcdMachineTemplate is the provider-specific machine template referenced
	// in EtcdCluster.Spec.InfrastructureTemplate
	EtcdMachineTemplate M
}

// Objects returns all API objects that form a concrete provider-specific control plane
func (cp *ControlPlane[C, M]) Objects() []kubernetes.Object {
	objs := make([]kubernetes.Object, 0, 4)
	objs = append(objs, cp.Cluster, cp.KubeadmControlPlane, cp.ProviderCluster, cp.ControlPlaneMachineTemplate)
	if cp.EtcdCluster != nil {
		objs = append(objs, cp.EtcdCluster, cp.EtcdMachineTemplate)
	}

	return objs
}

// UpdateImmutableObjectNames checks if any control plane immutable objects have changed by comparing the new definition
// with the current state of the cluster. If they had, it generates a new name for them by increasing a monotonic number
// at the end of the name
// This is applied to all provider machine templates
func (cp *ControlPlane[C, M]) UpdateImmutableObjectNames(
	ctx context.Context,
	client kubernetes.Client,
	machineTemplateRetriever ObjectRetriever[M],
	machineTemplateComparator ObjectComparator[M],
) error {
	currentKCP := &controlplanev1.KubeadmControlPlane{}
	err := client.Get(ctx, cp.KubeadmControlPlane.Name, cp.KubeadmControlPlane.Namespace, currentKCP)
	if apierrors.IsNotFound(err) {
		// KubeadmControlPlane doesn't exist, this is a new cluster so machine templates should use their default name
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "reading current kubeadm control plane from API")
	}

	cp.ControlPlaneMachineTemplate.SetName(currentKCP.Spec.MachineTemplate.InfrastructureRef.Name)
	if err = EnsureNewNameIfChanged(ctx, client, machineTemplateRetriever, machineTemplateComparator, cp.ControlPlaneMachineTemplate); err != nil {
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
		return errors.Wrap(err, "reading current etcdadm cluster from API")
	}

	cp.EtcdMachineTemplate.SetName(currentEtcdCluster.Spec.InfrastructureTemplate.Name)
	if err = EnsureNewNameIfChanged(ctx, client, machineTemplateRetriever, machineTemplateComparator, cp.EtcdMachineTemplate); err != nil {
		return err
	}

	return nil
}

// EnsureNewNameIfChanged updates an object's name if such object is different from its current state in the cluster
func EnsureNewNameIfChanged[M Object](ctx context.Context,
	client kubernetes.Client,
	retrieve ObjectRetriever[M],
	equal ObjectComparator[M],
	new M,
) error {
	current, err := retrieve(ctx, client, new.GetName(), new.GetNamespace())
	if apierrors.IsNotFound(err) {
		// if object doesn't exist with same name in same namespace, no need to compare, there won't be a conflict
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "reading %s %s/%s from API",
			new.GetObjectKind().GroupVersionKind().Kind,
			new.GetNamespace(),
			new.GetName(),
		)
	}

	if !equal(new, current) {
		newName, err := IncrementName(new.GetName())
		if err != nil {
			return errors.Wrapf(err, "incrementing name for %s %s/%s",
				new.GetObjectKind().GroupVersionKind().Kind,
				new.GetNamespace(),
				new.GetName(),
			)
		}

		new.SetName(newName)
	}

	return nil
}
