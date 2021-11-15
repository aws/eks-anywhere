package controllers

import (
	"context"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clust "github.com/aws/eks-anywhere/pkg/cluster"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// TODO: find a better name, possibly better pattern -> group with something else
func (r *ClusterReconcilerV2) reconcileExtraObjects(ctx context.Context, cluster *anywhere.Cluster) error {
	spec, err := r.Client.BuildClusterSpec(ctx, cluster)
	if err != nil {
		return err
	}

	extraObjects := clust.BuildExtraObjects(spec)

	remoteClient, err := r.BuildRemoteClientForCluster(ctx, cluster)
	if err != nil {
		return err
	}

	for _, obj := range extraObjects {
		objs, err := yamlToUnstructured(obj)
		if err != nil {
			return err
		}

		for _, o := range objs {
			if err := remoteClient.Create(ctx, &o); err != nil {
				if apierrors.IsAlreadyExists(err) {
					// TODO: update if necessary
					continue
				}
				return err
			}
		}
	}

	return nil
}
