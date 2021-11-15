package controllers

import (
	"context"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	capicontrolplane "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
)

const ClusterLabelName = "cluster.anywhere.eks.amazonaws.com/cluster-name"

func (r *ClusterReconcilerV2) reconcileControlPlane(ctx context.Context, cluster *anywhere.Cluster) error {
	// TODO: find a way to support spec changes reconciliation. In order to check if something has changed,
	//  the only way is to check object by object if the objects coming from the "template" are the different than the ones already present in the cluster.
	//  This is an expensive operation so we could potentially cache it in the status of our cluster.
	//  We would need a combination of all the observed generation for all the resources than can affect
	//  such objects (cluster itself, datacenter, machineconfigs...) and the "phase" (in this case, "CP reconciled")
	//  Right now, object comparison might be a pain, bc the providers generate multiobject yaml files. We might be able to get around this by
	//  by coverting first to unstructured and then using kind, name and namespace to get unstructured objects from the api server. Then use some kind of
	//  semantic Equal comparison (I think one of the api kubernetes libraries provides this already)
	//  In the future, this will be easier if we start using api structs in the providers instead of straight yaml
	// For now, I'm only going to check if the CAPI cluster exists
	// If it doesn't, run the whole thing and create all objects blindly, if it does, return immediately
	// In order to check if the CAPI cluster exists, I use a label (which need to be applied on creation)
	// We can "cache" this search by storing the capi cluster name in our cluster status
	capiCluster, err := r.Client.GetCAPICluster(ctx, cluster)
	if err != nil {
		return err
	}

	if capiCluster != nil {
		r.Log.Info("CAPI cluster already exists, skipping CP reconcile", "cluster", cluster.Name)
		return nil
	}

	r.Log.Info("Reconciling control plane", "cluster", cluster.Name)
	return r.buildProviderReconciler(cluster).ReconcileControlPlane(ctx, cluster)
}

func capiControlPlaneReady(capiControlPlane *capicontrolplane.KubeadmControlPlane) bool {
	status := capiControlPlane.Status
	return status.Ready && status.UnavailableReplicas == 0 && status.ReadyReplicas == status.Replicas
}
