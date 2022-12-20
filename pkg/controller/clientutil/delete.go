package clientutil

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteYaml(ctx context.Context, c client.Client, yaml []byte) error {
	objs, err := YamlToClientObjects(yaml)
	if err != nil {
		return err
	}

	return deleteObjects(ctx, c, objs)
}

func deleteObjects(ctx context.Context, c client.Client, objs []client.Object) error {
	for _, o := range objs {
		if err := deleteObject(ctx, c, o); err != nil {
			return err
		}
	}

	return nil
}

func deleteObject(ctx context.Context, c client.Client, obj client.Object) error {
	if err := c.Delete(ctx, obj); err != nil {
		return errors.Wrapf(err, "deleting object %s, %s/%s", obj.GetObjectKind().GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	}

	return nil
}
