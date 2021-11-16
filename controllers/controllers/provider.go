package controllers

import (
	"context"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type providerReconciler interface {
	ReconcileControlPlane(ctx context.Context, cluster *anywhere.Cluster) error
	ReconcileWorkers(ctx context.Context, cluster *anywhere.Cluster) error
}

func (r *ClusterReconcilerV2) buildProviderReconciler(cluster *anywhere.Cluster) providerReconciler {
	switch cluster.Spec.DatacenterRef.Kind {
	case anywhere.VSphereDatacenterKind:
		return &vsphere{client: r.Client}
	case anywhere.DockerDatacenterKind:
		return &docker{client: r.Client}
	default:
		// TODO: handle this kind of error?
		return nil
	}
}
