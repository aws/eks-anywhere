package serverside

import (
	"context"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

const fieldManager = "eks-a-controller"

func ReconcileYaml(ctx context.Context, c client.Client, yaml []byte) error {
	objs, err := clientutil.YamlToClientObjects(yaml)
	if err != nil {
		return err
	}

	return ReconcileObjects(ctx, c, objs)
}

func ReconcileObjects(ctx context.Context, c client.Client, objs []client.Object) error {
	for _, o := range objs {
		if err := ReconcileObject(ctx, c, o); err != nil {
			return err
		}
	}

	return nil
}

func ReconcileObject(ctx context.Context, c client.Client, obj client.Object) error {
	// Server side apply
	err := c.Patch(ctx, obj, client.Apply, client.FieldOwner(fieldManager), client.ForceOwnership)
	if err != nil {
		return errors.Wrapf(err, "failed to reconcile object %s, %s/%s", obj.GetObjectKind().GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}

	return nil
}

// UpdateObject updates the existing object during reconciliation.
// This is intended for special use cases only as the preferred method to reconcile objects is server-side apply.
func UpdateObject(ctx context.Context, c client.Client, obj client.Object) error {
	if err := c.Update(ctx, obj, client.FieldOwner(fieldManager)); err != nil {
		return errors.Wrapf(err, "failed to reconcile object %s, %s/%s", obj.GetObjectKind().GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}

	return nil
}

// DeleteYaml deletes Kubernetes objects from YAML content.
func DeleteYaml(ctx context.Context, c client.Client, yaml []byte) error {
	objs, err := clientutil.YamlToClientObjects(yaml)
	if err != nil {
		return err
	}

	return DeleteObjects(ctx, c, objs)
}

// DeleteObjects deletes multiple Kubernetes objects.
func DeleteObjects(ctx context.Context, c client.Client, objs []client.Object) error {
	for _, o := range objs {
		if err := DeleteObject(ctx, c, o); err != nil {
			return err
		}
	}

	return nil
}

// DeleteObject deletes a single Kubernetes object. It's idempotent - if the object doesn't exist, no error is returned.
func DeleteObject(ctx context.Context, c client.Client, obj client.Object) error {
	err := c.Delete(ctx, obj)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete object %s, %s/%s", obj.GetObjectKind().GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}

	return nil
}
