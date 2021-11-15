package controllers

import (
	"context"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ClusterReconcilerV2) reconcileCNI(ctx context.Context, cluster *anywhere.Cluster) error {
	// TODO: support CNI true reconciliation, this only support installing it once

	remoteClient, err := r.BuildRemoteClientForCluster(ctx, cluster)
	if err != nil {
		return err
	}

	// This is a bit hacky, just check if the DS for cilium exists
	ds := &v1.DaemonSet{}
	if err := remoteClient.Get(ctx, types.NamespacedName{Name: "cilium", Namespace: "kube-system"}, ds); err == nil {
		r.Log.Info("Cilium DaemonSet already exists, skipping CNI reconcile")
		return nil
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	spec, err := r.Client.BuildClusterSpec(ctx, cluster)
	if err != nil {
		return err
	}

	ciliumYaml, err := spec.LoadManifest(spec.VersionsBundle.Cilium.Manifest)
	if err != nil {
		return err
	}

	// Convert cilium yaml manifest to objs
	objs, err := yamlToUnstructured(ciliumYaml.Content)
	if err != nil {
		return err
	}

	r.Log.Info("Reconciling Cilium CNI")
	for _, obj := range objs {
		if err := remoteClient.Create(ctx, &obj); err != nil {
			if apierrors.IsAlreadyExists(err) {
				r.Log.Info("Cilium obj already exists, this should not happen", "objName", obj.GetName())
			}
			return err
		}
	}

	return nil
}
