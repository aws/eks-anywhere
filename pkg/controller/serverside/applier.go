package serverside

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

type ObjectGenerator func() ([]kubernetes.Object, error)

// ObjectApplier helps reconcile kubernetes object using server side apply.
type ObjectApplier struct {
	client client.Client
}

// NewObjectApplier builds a ObjectApplier.
func NewObjectApplier(client client.Client) *ObjectApplier {
	return &ObjectApplier{
		client: client,
	}
}

// Apply uses server side apply to reconcile kubernetes objects returned by a generator
// Useful in reconcilers because it simplifies the reconciliation when generating API
// objects from another package, like a provider
// This is mostly a helper for generate objects + serverside apply.
func (a *ObjectApplier) Apply(ctx context.Context, generateObjects ObjectGenerator) (controller.Result, error) {
	return controller.Result{}, reconcileKubernetesObjects(ctx, a.client, generateObjects)
}

func reconcileKubernetesObjects(ctx context.Context, client client.Client, generateObjects ObjectGenerator) error {
	objs, err := generateObjects()
	if err != nil {
		return err
	}

	if err = ReconcileObjects(ctx, client, clientutil.ObjectsToClientObjects(objs)); err != nil {
		return err
	}

	return nil
}
