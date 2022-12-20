package clusters_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestProviderClusterReconcilerRegistryGet(t *testing.T) {
	dummy1 := dummyProviderReconciler{name: "dummy1"}
	dummy2 := dummyProviderReconciler{name: "dummy2"}
	registry := clusters.NewProviderClusterReconcilerRegistryBuilder().
		Add("dummy1", dummy1).
		Add("dummy2", dummy2).
		Build()

	tests := []struct {
		name           string
		datacenterKind string
		want           clusters.ProviderClusterReconciler
	}{
		{
			name:           "reconciler exists",
			datacenterKind: "dummy1",
			want:           dummy1,
		},
		{
			name:           "reconciler does not exist",
			datacenterKind: "dummy3",
			want:           nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			if tt.want != nil {
				g.Expect(registry.Get(tt.datacenterKind)).To(Equal(tt.want))
			} else {
				g.Expect(registry.Get(tt.datacenterKind)).To(BeNil())
			}
		})
	}
}

type dummyProviderReconciler struct {
	name string
}

func (dummyProviderReconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}

func (dummyProviderReconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	return controller.Result{}, nil
}

func (dummyProviderReconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}
