package envtest

import (
	"context"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object interface {
	client.Object
	schema.ObjectKind
}

// CreateObjs creates Objects using the provided kube client and waits until its cache
// has been updated with those objects
func CreateObjs(ctx context.Context, t testing.TB, client client.Client, objs ...Object) {
	t.Helper()
	for _, obj := range objs {
		// client.Create cleans the group version and kind of the objects
		// We just save it before calling and it and restore it later,
		// just so we can use it to perform the get later
		objKind := obj.GroupVersionKind()
		if err := client.Create(ctx, obj); err != nil {
			t.Fatal(err)
		}
		obj.SetGroupVersionKind(objKind)
	}

	for _, obj := range objs {
		for {
			o := &unstructured.Unstructured{}
			o.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
			if err := client.Get(ctx, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, o); err == nil {
				break
			} else if !apierrors.IsNotFound(err) {
				t.Fatal(err)
			}
		}
	}
}
