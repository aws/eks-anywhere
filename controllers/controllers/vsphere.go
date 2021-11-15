package controllers

import (
	"context"

	anywhere "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type vsphere struct{}

func (v *vsphere) ReconcileControlPlane(ctx context.Context, cluster *anywhere.Cluster) error {
	return nil
}

func (v *vsphere) ReconcileWorkers(ctx context.Context, cluster *anywhere.Cluster) error {
	return nil
}
