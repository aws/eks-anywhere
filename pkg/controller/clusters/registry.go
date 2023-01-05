package clusters

import (
	"context"

	"github.com/go-logr/logr"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

// ProviderClusterReconciler reconciles a provider specific eks-a cluster.
type ProviderClusterReconciler interface {
	// Reconcile handles the full cluster reconciliation.
	Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
	// ReconcileWorkerNodes handles only the worker node reconciliation. Intended to be used on self managed clusters.
	ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
}

// ProviderClusterReconcilerRegistry holds a collection of cluster provider reconcilers
// and ties them to different provider Datacenter kinds.
type ProviderClusterReconcilerRegistry struct {
	reconcilers map[string]ProviderClusterReconciler
}

func newClusterReconcilerRegistry() ProviderClusterReconcilerRegistry {
	return ProviderClusterReconcilerRegistry{
		reconcilers: map[string]ProviderClusterReconciler{},
	}
}

func (r *ProviderClusterReconcilerRegistry) add(datacenterKind string, reconciler ProviderClusterReconciler) {
	r.reconcilers[datacenterKind] = reconciler
}

// Get returns ProviderClusterReconciler for a particular Datacenter kind.
func (r *ProviderClusterReconcilerRegistry) Get(datacenterKind string) ProviderClusterReconciler {
	return r.reconcilers[datacenterKind]
}

// ProviderClusterReconcilerRegistryBuilder builds ProviderClusterReconcilerRegistry's.
type ProviderClusterReconcilerRegistryBuilder struct {
	reconciler ProviderClusterReconcilerRegistry
}

// NewProviderClusterReconcilerRegistryBuilder returns a new empty ProviderClusterReconcilerRegistryBuilder.
func NewProviderClusterReconcilerRegistryBuilder() *ProviderClusterReconcilerRegistryBuilder {
	return &ProviderClusterReconcilerRegistryBuilder{
		reconciler: newClusterReconcilerRegistry(),
	}
}

// Add accumulates a pair of datacenter kind a reconciler to be included in the final registry.
func (b *ProviderClusterReconcilerRegistryBuilder) Add(datacenterKind string, reconciler ProviderClusterReconciler) *ProviderClusterReconcilerRegistryBuilder {
	b.reconciler.add(datacenterKind, reconciler)
	return b
}

// Build returns a registry with all the previously added reconcilers.
func (b *ProviderClusterReconcilerRegistryBuilder) Build() ProviderClusterReconcilerRegistry {
	r := b.reconciler
	b.reconciler = newClusterReconcilerRegistry()
	return r
}
