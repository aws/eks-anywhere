package serverside

import (
	"context"
	"fmt"

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
	if objs == nil {
		fmt.Println("[DEBUG] ReconcileObjects: objs array is nil!")
	}
	for _, o := range objs {
		if o == nil {
			return fmt.Errorf("object in array is nil. Full array: %s", objs)
		}
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
		fmt.Println("[DEBUG] failed to reconcile object", obj)
		return errors.Wrapf(err, "failed to reconcile object %s, %s/%s", obj.GetObjectKind().GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}
	fmt.Println("[DEBUG] Successfully reconciled obj", obj.GetName())

	return nil
}
