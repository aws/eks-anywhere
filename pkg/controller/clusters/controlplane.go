package clusters

import (
	"context"
	"reflect"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
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
	objs := cp.nonEtcdObjects()
	if cp.EtcdCluster != nil {
		objs = append(objs, cp.etcdObjects()...)
	}

	return objs
}

func (cp *ControlPlane) etcdObjects() []client.Object {
	return []client.Object{cp.EtcdMachineTemplate, cp.EtcdCluster}
}

func (cp *ControlPlane) nonEtcdObjects() []client.Object {
	objs := make([]client.Object, 0, 4+len(cp.Other))
	objs = append(objs, cp.Cluster, cp.ProviderCluster, cp.KubeadmControlPlane)
	if !reflect.ValueOf(cp.ControlPlaneMachineTemplate).IsNil() {
		objs = append(objs, cp.ControlPlaneMachineTemplate)
	}
	objs = append(objs, cp.Other...)

	return objs
}

// skipCAPIAutoPauseKCPForExternalEtcdAnnotation instructs the CAPI cluster controller to not pause or unpause
// the KCP to wait for etcd endpoints to be ready. When this annotation is present, is left to the user (us)
// to orchestrate this operation if double kcp rollouts are undesirable.
const skipCAPIAutoPauseKCPForExternalEtcdAnnotation = "cluster.x-k8s.io/skip-pause-cp-managed-etcd"

// ReconcileControlPlane orchestrates the ControlPlane reconciliation logic.
func ReconcileControlPlane(ctx context.Context, log logr.Logger, c client.Client, cp *ControlPlane) (controller.Result, error) {
	if cp.EtcdCluster == nil {
		// For stacked etcd, we don't need orchestration, apply directly
		return controller.Result{}, applyAllControlPlaneObjects(ctx, c, cp)
	}

	// always add skip pause annotation since we want to have full control over the kcp-etcd orchestration
	clientutil.AddAnnotation(cp.KubeadmControlPlane, skipCAPIAutoPauseKCPForExternalEtcdAnnotation, "true")

	cluster := &clusterv1.Cluster{}
	err := c.Get(ctx, client.ObjectKeyFromObject(cp.Cluster), cluster)
	if apierrors.IsNotFound(err) {
		log.Info("Creating cluster with external etcd")
		// If the CAPI cluster doesn't exist, this is a new cluster, create all objects
		return controller.Result{}, applyAllControlPlaneObjects(ctx, c, cp)
	}
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "reading CAPI cluster")
	}

	etcdadmCluster, err := getEtcdadmCluster(ctx, c, cluster)
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "reading CAPI cluster")
	}

	kcp := &controlplanev1.KubeadmControlPlane{}
	if err = c.Get(ctx, objKeyForRef(cluster.Spec.ControlPlaneRef), kcp); err != nil {
		return controller.Result{}, errors.Wrap(err, "reading kubeadm control plane")
	}

	// If there are changes for etcd, we only apply those changes for now and we wait.
	if !equality.Semantic.DeepDerivative(cp.EtcdCluster.Spec, etcdadmCluster.Spec) {
		return reconcileEtcdChanges(ctx, log, c, cp, kcp, etcdadmCluster)
	}

	// If etcd is not ready yet, we requeue and wait before making any other change to the control plane.
	if !etcdadmClusterReady(etcdadmCluster) {
		// We need to inject a logger in this method or extract from context
		log.Info("Etcd is not ready, requeuing")
		return controller.ResultWithRequeue(30 * time.Second), nil
	}

	return reconcileControlPlaneNodeChanges(ctx, log, c, cp, kcp)
}

func applyAllControlPlaneObjects(ctx context.Context, c client.Client, cp *ControlPlane) error {
	if err := serverside.ReconcileObjects(ctx, c, cp.AllObjects()); err != nil {
		return errors.Wrap(err, "applying control plane objects")
	}

	return nil
}

func reconcileEtcdChanges(ctx context.Context, log logr.Logger, c client.Client, desiredCP *ControlPlane, currentKCP *controlplanev1.KubeadmControlPlane, currentEtcdadmCluster *etcdv1.EtcdadmCluster) (controller.Result, error) {
	// Before making any changes to etcd, pause the KCP so it doesn't rollout new nodes as the
	// etcd endpoints change.
	if !annotations.HasPaused(currentKCP) {
		log.Info("Pausing KCP before making any etcd changes", "kcp", klog.KObj(currentKCP))
		clientutil.AddAnnotation(currentKCP, clusterv1.PausedAnnotation, "true")
		if err := c.Update(ctx, currentKCP); err != nil {
			return controller.Result{}, err
		}
	}

	// If the etcdadm cluster has changes, this will require a rolling upgrade
	// Mark the etcdadm cluster as upgrading
	// The etcdadm cluster controller will take care of removing
	// this annotation at the right time to orchestrate the kcp upgrade.
	clientutil.AddAnnotation(desiredCP.EtcdCluster, etcdv1.UpgradeInProgressAnnotation, "true")

	log.Info("Reconciling external etcd changes", "etcdadmCluster", klog.KObj(currentEtcdadmCluster))
	if err := serverside.ReconcileObjects(ctx, c, desiredCP.etcdObjects()); err != nil {
		return controller.Result{}, errors.Wrap(err, "applying etcd objects")
	}

	// After applying etcd changes, just requeue to wait until etcd finishes updating and is ready
	// We use a short wait here just in case etcdadm controller decides that not new machines are
	// needed.
	log.Info("Requeuing to wait until etcd is ready")
	return controller.ResultWithRequeue(10 * time.Second), nil
}

func etcdadmClusterReady(etcdadmCluster *etcdv1.EtcdadmCluster) bool {
	// It's important to use status.Ready and not the Ready condition, since the Ready condition
	// only becomes true after the old etcd members have been deleted, which only happens after the
	// kcp finishes its own upgrade.
	return etcdadmCluster.Generation == etcdadmCluster.Status.ObservedGeneration && etcdadmCluster.Status.Ready
}

func reconcileControlPlaneNodeChanges(ctx context.Context, log logr.Logger, c client.Client, desiredCP *ControlPlane, currentKCP *controlplanev1.KubeadmControlPlane) (controller.Result, error) {
	// When the controller reconciles the control plane for a cluster with an external etcd configuration
	// the KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints field is
	// defaulted to an empty slice. However, at some point that field in KubeadmControlPlane object is filled
	// and updated by the kcp controller.
	//
	// We do not want to update the field with an empty slice again, so here we check if the endpoints for the
	// external etcd have already been populated on the KubeadmControlPlane object and override ours before applying it.
	externalEndpoints := currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints
	if len(externalEndpoints) != 0 {
		desiredCP.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = externalEndpoints
	}

	if err := serverside.ReconcileObjects(ctx, c, desiredCP.nonEtcdObjects()); err != nil {
		return controller.Result{}, errors.Wrap(err, "applying non etcd control plane objects")
	}

	// If the KCP is paused, we read the last version (in case we just updated it) and unpause it
	// so the cp nodes are reconciled.
	if annotations.HasPaused(currentKCP) {
		kcp := &controlplanev1.KubeadmControlPlane{}
		if err := c.Get(ctx, client.ObjectKeyFromObject(currentKCP), kcp); err != nil {
			return controller.Result{}, errors.Wrap(err, "reading updates kubeadm control plane to unpause")
		}

		delete(kcp.Annotations, clusterv1.PausedAnnotation)
		log.Info("Unpausing KCP after update to start reconciliation", klog.KObj(currentKCP))
		if err := c.Update(ctx, kcp); err != nil {
			return controller.Result{}, err
		}
	}

	return controller.Result{}, nil
}

func getEtcdadmCluster(ctx context.Context, c client.Client, cluster *clusterv1.Cluster) (*etcdv1.EtcdadmCluster, error) {
	key := objKeyForRef(cluster.Spec.ManagedExternalEtcdRef)
	// This can happen when a user has a workload cluster that is older than the following PR, causing cluster
	// reconcilation to fail. By inferring namespace from clusterv1.Cluster, we will be able to retrieve the object correctly.
	// PR: https://github.com/aws/eks-anywhere/pull/4025
	// TODO: See if it is possible to propagate the namespace field in the clusterv1.Cluster object in cluster-api like the other refs.
	if key.Namespace == "" {
		key.Namespace = cluster.Namespace
	}

	etcdadmCluster := &etcdv1.EtcdadmCluster{}
	if err := c.Get(ctx, key, etcdadmCluster); err != nil {
		return nil, errors.Wrap(err, "reading etcdadm cluster")
	}

	return etcdadmCluster, nil
}

func objKeyForRef(ref *corev1.ObjectReference) client.ObjectKey {
	return client.ObjectKey{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
}
