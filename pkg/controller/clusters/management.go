package clusters

import (
	"context"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// FetchManagementEksaCluster returns the management cluster object for a given workload Cluster.
// If we are unable to find the management cluster using the same namespace as the current cluster, we will attempt
// to get a list of the clusters with that name across all the namespaces. If we find multiple, which usually should
// not happen as these clusters get mapped to a cluster-api cluster object in the eksa-system namespace, then we
// also error on that because it is not possible to have multiple resources with the same name within a namespace.
func FetchManagementEksaCluster(ctx context.Context, cli client.Client, cluster *v1alpha1.Cluster) (*v1alpha1.Cluster, error) {
	mgmtCluster := &v1alpha1.Cluster{}
	mgmtClusterKey := client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cluster.ManagedBy(),
	}
	err := cli.Get(ctx, mgmtClusterKey, mgmtCluster)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if apierrors.IsNotFound(err) {
		// Save error returned from Get if we don't end up finding the cluster through List as it won't return an error
		notFoundErr := errors.Wrapf(err, "unable to retrieve management cluster %s", cluster.Spec.ManagementCluster.Name)
		clusterList := &v1alpha1.ClusterList{}
		if err = cli.List(ctx, clusterList,
			client.MatchingFields{"metadata.name": cluster.Spec.ManagementCluster.Name}); err != nil {
			return nil, errors.Wrapf(err, "unable to retrieve management cluster %s", cluster.Spec.ManagementCluster.Name)
		}
		if len(clusterList.Items) == 0 {
			return nil, notFoundErr
		}
		if len(clusterList.Items) > 1 {
			return nil, errors.Errorf("found multiple clusters with the name %s", cluster.Spec.ManagementCluster.Name)
		}
		mgmtCluster = &clusterList.Items[0]
	}
	return mgmtCluster, nil
}
