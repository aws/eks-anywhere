package clusters

import (
	"context"
	"reflect"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
)

// ControlPlane represents a CAPI spec for a kubernetes cluster.
type ControlPlane struct {
	Cluster *clusterv1.Cluster

	// ProviderCluster is the provider-specific resource that holds the details
	// for provisioning the infrastructure, referenced in Cluster.Spec.InfrastructureRef
	ProviderCluster client.Object

	KubeadmControlPlane *controlplanev1.KubeadmControlPlane

	// ControlPlaneMachineTemplate is the provider-specific machine template referenced
	// in KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef
	ControlPlaneMachineTemplate client.Object

	EtcdCluster *etcdv1.EtcdadmCluster

	// EtcdMachineTemplate is the provider-specific machine template referenced
	// in EtcdCluster.Spec.InfrastructureTemplate
	EtcdMachineTemplate client.Object

	// Other includes any other provider-specific objects that need to be reconciled
	// as part of the control plane.
	Other []client.Object
}

// AllObjects returns all the control plane objects.
func (cp *ControlPlane) AllObjects() []client.Object {
	objs := make([]client.Object, 0, 6+len(cp.Other))
	objs = append(objs, cp.Cluster, cp.ProviderCluster, cp.KubeadmControlPlane)
	if !reflect.ValueOf(cp.ControlPlaneMachineTemplate).IsNil() {
		objs = append(objs, cp.ControlPlaneMachineTemplate)
	}
	if cp.EtcdCluster != nil {
		objs = append(objs, cp.EtcdCluster, cp.EtcdMachineTemplate)
	}
	objs = append(objs, cp.Other...)

	return objs
}

// ReconcileControlPlane orchestrates the ControlPlane reconciliation logic.
func ReconcileControlPlane(ctx context.Context, c client.Client, cp *ControlPlane) (controller.Result, error) {
	if cp.EtcdCluster == nil {
		// For stacked etcd, we don't need orchestration, apply directly
		return controller.Result{}, applyAllControlPlaneObjects(ctx, c, cp)
	}

	cluster := &clusterv1.Cluster{}
	err := c.Get(ctx, client.ObjectKeyFromObject(cp.Cluster), cluster)
	if apierrors.IsNotFound(err) {
		// If the CAPI cluster doesn't exist, this is a new cluster, create all objects
		return controller.Result{}, applyAllControlPlaneObjects(ctx, c, cp)
	}
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "reading CAPI cluster")
	}

	externalEtcdNamespace := cluster.Spec.ManagedExternalEtcdRef.Namespace
	// This can happen when a user has a workload cluster that is older than the following PR, causing cluster
	// reconcilation to fail. By inferring namespace from clusterv1.Cluster, we will be able to retrieve the object correctly.
	// PR: https://github.com/aws/eks-anywhere/pull/4025
	// TODO: See if it is possible to propagate the namespace field in the clusterv1.Cluster object in cluster-api like the other refs.
	if externalEtcdNamespace == "" {
		externalEtcdNamespace = cluster.Namespace
	}

	etcdadmCluster := &etcdv1.EtcdadmCluster{}
	key := client.ObjectKey{
		Name:      cluster.Spec.ManagedExternalEtcdRef.Name,
		Namespace: externalEtcdNamespace,
	}
	if err = c.Get(ctx, key, etcdadmCluster); err != nil {
		return controller.Result{}, errors.Wrap(err, "reading etcdadm cluster")
	}

	if !equality.Semantic.DeepDerivative(cp.EtcdCluster.Spec, etcdadmCluster.Spec) {
		// If the etcdadm cluster has changes, this will require a rolling upgrade
		// Mark the etcdadm cluster as upgrading and pause the kcp reconciliation
		// The CAPI cluster and etcdadm cluster controller will take care of removing
		// these annotation at the right time to orchestrate the kcp upgrade
		clientutil.AddAnnotation(cp.EtcdCluster, etcdv1.UpgradeInProgressAnnotation, "true")
		clientutil.AddAnnotation(cp.KubeadmControlPlane, clusterv1.PausedAnnotation, "true")
	}

	return controller.Result{}, applyAllControlPlaneObjects(ctx, c, cp)
}

func applyAllControlPlaneObjects(ctx context.Context, c client.Client, cp *ControlPlane) error {
	if err := serverside.ReconcileObjects(ctx, c, cp.AllObjects()); err != nil {
		return errors.Wrap(err, "applying control plane objects")
	}

	return nil
}
