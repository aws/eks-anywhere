package clusters

import (
	"context"
	"reflect"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
)

// ControlPlane represents a CAPI spec for a kubernetes cluster.
type ControlPlane struct {
	Cluster *clusterv1beta2.Cluster

	// ProviderCluster is the provider-specific resource that holds the details
	// for provisioning the infrastructure, referenced in Cluster.Spec.InfrastructureRef
	ProviderCluster client.Object

	KubeadmControlPlane *controlplanev1beta2.KubeadmControlPlane

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
	cluster, kcp, etcdadmCluster, err := readCurrentControlPlane(ctx, c, cp)
	if err != nil {
		return controller.Result{}, err
	}

	if cp.EtcdCluster != nil {
		// always add skip pause annotation since we want to have full control over the kcp-etcd orchestration
		clientutil.AddAnnotation(cp.KubeadmControlPlane, skipCAPIAutoPauseKCPForExternalEtcdAnnotation, "true")
	}

	if cluster == nil {
		log.Info("Creating cluster")
		// If the CAPI cluster doesn't exist, this is a new cluster, create all objects, no need for extra orchestration.
		return controller.Result{}, applyAllControlPlaneObjects(ctx, c, cp)
	}

	currentCPEndpoint := cluster.Spec.ControlPlaneEndpoint
	desiredCPEndpoint := cp.Cluster.Spec.ControlPlaneEndpoint
	if desiredCPEndpoint.IsZero() && !currentCPEndpoint.IsZero() {
		// If the control plane endpoint is not set in the desired cluster, we want to keep the current one.
		// In practice, this condition will always be hit because:
		// * We don't set the endpoint in the cluster object in our code, we let CAPI do that
		// * The endpoint never changes once the cluster has been created
		cp.Cluster.Spec.ControlPlaneEndpoint = currentCPEndpoint
	}

	if cp.EtcdCluster == nil {
		// For stacked etcd, we don't need orchestration, apply directly
		return controller.Result{}, applyAllControlPlaneObjects(ctx, c, cp)
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

func readCurrentControlPlane(ctx context.Context, c client.Client, cp *ControlPlane) (*clusterv1beta2.Cluster, *controlplanev1beta2.KubeadmControlPlane, *etcdv1.EtcdadmCluster, error) {
	cluster := &clusterv1beta2.Cluster{}
	err := c.Get(ctx, client.ObjectKeyFromObject(cp.Cluster), cluster)
	if apierrors.IsNotFound(err) {
		// If the CAPI cluster doesn't exist, this is a new cluster, no need to read the rest of the objects.
		return nil, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "reading CAPI cluster")
	}

	kcp := &controlplanev1beta2.KubeadmControlPlane{}
	key := client.ObjectKey{
		Name:      cluster.Spec.ControlPlaneRef.Name,
		Namespace: cluster.Namespace,
	}
	if err = c.Get(ctx, key, kcp); err != nil {
		return nil, nil, nil, errors.Wrap(err, "reading kubeadm control plane")
	}

	if cp.EtcdCluster == nil {
		return cluster, kcp, nil, nil
	}

	etcdadmCluster, err := getEtcdadmCluster(ctx, c, cluster)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "reading etcdadm cluster")
	}

	return cluster, kcp, etcdadmCluster, nil
}

func applyAllControlPlaneObjects(ctx context.Context, c client.Client, cp *ControlPlane) error {
	if err := serverside.ReconcileObjects(ctx, c, cp.AllObjects()); err != nil {
		return errors.Wrap(err, "applying control plane objects")
	}

	return nil
}

func reconcileEtcdChanges(ctx context.Context, log logr.Logger, c client.Client, desiredCP *ControlPlane, currentKCP *controlplanev1beta2.KubeadmControlPlane, currentEtcdadmCluster *etcdv1.EtcdadmCluster) (controller.Result, error) {
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

func reconcileControlPlaneNodeChanges(ctx context.Context, log logr.Logger, c client.Client, desiredCP *ControlPlane, currentKCP *controlplanev1beta2.KubeadmControlPlane) (controller.Result, error) {
	// When the controller reconciles the control plane for a cluster with an external etcd configuration
	// the KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints field is
	// defaulted to a placeholder value. At some point that field in KubeadmControlPlane object is filled
	// and updated by the kcp controller with real etcd endpoints.
	//
	// We do not want to overwrite real endpoints with the placeholder again, so here we check if the endpoints
	// for the external etcd have already been populated with real values on the KubeadmControlPlane object
	// and preserve them in the desired state before applying.
	currentEndpoints := currentKCP.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints
	if len(currentEndpoints) != 0 && !isPlaceholderEndpoint(currentEndpoints) {
		desiredCP.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = currentEndpoints
	}

	if err := serverside.ReconcileObjects(ctx, c, desiredCP.nonEtcdObjects()); err != nil {
		return controller.Result{}, errors.Wrap(err, "applying non etcd control plane objects")
	}

	// If the KCP is paused, we read the last version (in case we just updated it) and unpause it
	// so the cp nodes are reconciled.
	if annotations.HasPaused(currentKCP) {
		kcp := &controlplanev1beta2.KubeadmControlPlane{}
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

// isPlaceholderEndpoint checks if endpoints are placeholder values that we set by default.
func isPlaceholderEndpoint(endpoints []string) bool {
	if len(endpoints) == 1 && endpoints[0] == "https://placeholder:2379" {
		return true
	}
	return false
}

func getEtcdadmCluster(ctx context.Context, c client.Client, cluster *clusterv1beta2.Cluster) (*etcdv1.EtcdadmCluster, error) {
	key := client.ObjectKey{
		Name:      cluster.Spec.ManagedExternalEtcdRef.Name,
		Namespace: cluster.Namespace,
	}

	etcdadmCluster := &etcdv1.EtcdadmCluster{}
	if err := c.Get(ctx, key, etcdadmCluster); err != nil {
		return nil, errors.Wrap(err, "reading etcdadm cluster")
	}

	return etcdadmCluster, nil
}
