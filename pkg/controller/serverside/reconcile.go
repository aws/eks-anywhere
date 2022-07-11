package serverside

import (
	"context"

	"github.com/pkg/errors"
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

func DeleteObjects(ctx context.Context, c client.Client, objs []client.Object) error {
	for _, o := range objs {
		if err := DeleteObject(ctx, c, o); err != nil {
			return err
		}
	}

	return nil
}

func DeleteObject(ctx context.Context, c client.Client, obj client.Object) error {
	// Server side apply
	err := c.Delete(ctx, obj)
	if err != nil {
		return errors.Wrapf(err, "failed to delete object %s, %s/%s", obj.GetObjectKind().GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}

	return nil
}

func ReconcileObject(ctx context.Context, c client.Client, obj client.Object) error {
	// Server side apply
	t := true
	err := c.Patch(ctx, obj, client.Apply, &client.PatchOptions{FieldManager: fieldManager, Force: &t})
	if err != nil {
		return errors.Wrapf(err, "failed to reconcile object %s, %s/%s", obj.GetObjectKind().GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}

	return nil
}

func DeleteYaml(ctx context.Context, c client.Client, yaml []byte) error {
	objs, err := clientutil.YamlToClientObjects(yaml)
	if err != nil {
		return err
	}

	return DeleteObjects(ctx, c, objs)
}
