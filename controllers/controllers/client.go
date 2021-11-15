package controllers

import (
	"context"
	"errors"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	spec "github.com/aws/eks-anywhere/pkg/cluster"
	anywhererelease "github.com/aws/eks-anywhere/release/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	capicontrolplane "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ClusterReconcilerV2) BuildRemoteClientForCluster(ctx context.Context, cluster *anywhere.Cluster) (*eksaClient, error) {
	capiCluster, err := r.Client.GetCAPICluster(ctx, cluster)
	if err != nil {
		return nil, err
	}
	key := client.ObjectKey{
		Namespace: capiCluster.Namespace,
		Name:      capiCluster.Name,
	}
	remoteClient, err := r.Tracker.GetClient(ctx, key)
	if err != nil {
		return nil, err
	}

	return &eksaClient{Client: remoteClient}, nil
}


type eksaClient struct {
	client.Client
}

func (c eksaClient) BuildClusterSpec(ctx context.Context, cluster *anywhere.Cluster) (*cluster.Spec, error) {
	bundles, err := c.GetBundlesForCluster(ctx, cluster)
	if err != nil {
		return nil, err
	}

	return spec.BuildSpecFromBundles(cluster, bundles)
}

func (c eksaClient) GetBundlesForCluster(ctx context.Context, cluster *anywhere.Cluster) (*anywhererelease.Bundles, error) {
	// TODO: the bundles name can be cached in the status
	// TODO: answer this question: do the management cluster and all the workload clusters use the same Bundle?
	//  if they don't, that opens a lot our testing matrix. Right now they do and they don't. The cli is not enforcing it,
	//  but the v0.6.0 is the only cli that supports workload clusters and there is only one bundle compatible with it
	// For now, I'm gonna get the bundles using the cluster name, but we should really change that

	// Once we have the bundles ref in the status, we can add a check here and use it if present
	// If the status is not there, it means it's a new cluster, so we get the bundles being used by the management cluster and use it
	// Maybe using the spec would be better than the status. Status is supposed to be recoverable from the state of the cluster,
	// and this seems more ephemeral
	// TODO: This assumed the Bundles object has the same name as the management cluster. Change when we have a ref
	// TODO: fix namespace issue, it seems like we install tge Bundles in the default namespace
	return c.GetBundles(ctx, cluster.Spec.ManagementCluster.Name, "default")
}

func (c eksaClient) GetBundles(ctx context.Context, name, namespace string) (*anywhererelease.Bundles, error) {
	bundles := &anywhererelease.Bundles{}
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, bundles); err != nil {
		return nil, err
	}
	return bundles, nil
}

func (c eksaClient) GetManagementCluster(ctx context.Context, cluster *anywhere.Cluster) (*anywhere.Cluster, error) {
	if cluster.IsSelfManaged() {
		return cluster, nil
	}

	managementCluster := &anywhere.Cluster{}
	if err := c.Get(ctx, types.NamespacedName{Name: cluster.Spec.ManagementCluster.Name, Namespace: cluster.Namespace}, managementCluster); err != nil {
		return nil, err
	}
	return managementCluster, nil
}

func (c eksaClient) GetCAPICluster(ctx context.Context, cluster *anywhere.Cluster) (*capi.Cluster, error) {
	capiClusterList := &capi.ClusterList{}
	if err := c.listObjectFromClusterLabelName(ctx, cluster.Name, cluster.Namespace, capiClusterList); err != nil {
		return nil, err
	}

	switch len(capiClusterList.Items) {
	case 2:
		return nil, errors.New("found more than one CAPI cluster with the cluster label name")
	case 0:
		return nil, nil
	}

	return &capiClusterList.Items[0], nil
}

func (c eksaClient) GetCAPIControlPlane(ctx context.Context, cluster *anywhere.Cluster) (*capicontrolplane.KubeadmControlPlane, error) {
	controlPlaneList := &capicontrolplane.KubeadmControlPlaneList{}
	if err := c.listObjectFromClusterLabelName(ctx, cluster.Name, cluster.Namespace, controlPlaneList); err != nil {
		return nil, err
	}

	// TODO: dry out this a bit more. Probably do some magic with interfaces
	switch len(controlPlaneList.Items) {
	case 2:
		return nil, errors.New("found more than one CAPI kubeadm control planes with the cluster label name")
	case 0:
		return nil, nil
	}

	return &controlPlaneList.Items[0], nil
}

func (c eksaClient) listObjectFromClusterLabelName(ctx context.Context, clusterName, namespace string, objList client.ObjectList) error {
	selectors := []client.ListOption{client.InNamespace(namespace), client.MatchingLabels{ClusterLabelName: clusterName}, client.Limit(2)}
	if err := c.List(ctx, objList, selectors...); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func (c eksaClient) GetMachineDeploymentsForMachineGroup(ctx context.Context, cluster *anywhere.Cluster, machineGroupName string) (*capi.MachineDeploymentList, error) {
	machineDeploymentList := &capi.MachineDeploymentList{}
	selectors := []client.ListOption{
		client.InNamespace(cluster.Namespace),
		client.MatchingLabels{ClusterLabelName: cluster.Name, MachineGroupLabelName: machineGroupName},
	}
	if err := c.List(ctx, machineDeploymentList, selectors...); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return machineDeploymentList, nil
}

func (c eksaClient) GetWorkerMachineDeployments(ctx context.Context, cluster *anywhere.Cluster) (*capi.MachineDeploymentList, error) {
	machineDeploymentList := &capi.MachineDeploymentList{}
	selectors := []client.ListOption{
		client.InNamespace(cluster.Namespace),
		client.MatchingLabels{ClusterLabelName: cluster.Name, MachineDeploymentLabelType: MachineDeploymentWorkersType},
	}
	if err := c.List(ctx, machineDeploymentList, selectors...); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return machineDeploymentList, nil
}
