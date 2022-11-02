package envtest

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Object client.Object

// CreateObjs creates Objects using the provided kube client and waits until its cache
// has been updated with those objects.
func CreateObjs(ctx context.Context, t testing.TB, c client.Client, objs ...Object) {
	t.Helper()
	for _, o := range objs {
		// we copy objects because the client modifies them while making creating/updating calls
		obj := copyObject(t, o)

		if err := c.Create(ctx, obj); isNamespace(obj) && apierrors.IsAlreadyExists(err) {
			// namespaces can't be deleted
			// assuming most tests just want the namespace to exist, since it already does
			// we ignore the error
			// for more advance usecases, handle namespaces manually outside of this helper
			continue
		} else if err != nil {
			t.Fatal(err)
		}
	}

	for _, o := range objs {
		obj := copyObject(t, o)
		objReady := waitForObjectReady(ctx, t, c, obj)

		// We need to update the status independently, kubernetes doesn't allow to create the main objects and
		// its subresources all at once
		obj.SetResourceVersion(objReady.GetResourceVersion())
		if err := c.Status().Update(ctx, obj); apierrors.IsNotFound(err) {
			// There is not easy way to check if an object has a status subresource
			// So we just try and if it fails with a 404, we ignore the error
			t.Logf(
				"Try updating status but failed with a 404 error for [%s name=%s namespace=%s] object, most probably because it doesn't have a defined status subresource",
				obj.GetObjectKind().GroupVersionKind().String(),
				obj.GetName(),
				obj.GetNamespace(),
			)
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func waitForObjectReady(ctx context.Context, t testing.TB, c client.Client, obj Object) Object {
	unstructuredObj := &unstructured.Unstructured{}
	for {
		unstructuredObj.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		if err := c.Get(ctx, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, unstructuredObj); err == nil {
			break
		} else if !apierrors.IsNotFound(err) {
			t.Fatal(err)
		}
	}

	return unstructuredObj
}

func isNamespace(obj Object) bool {
	_, isNamespace := obj.(*v1.Namespace)
	return isNamespace
}

func copyObject(t testing.TB, obj Object) Object {
	copyRuntimeObj := obj.DeepCopyObject()
	copyObj, ok := copyRuntimeObj.(Object)
	if !ok {
		t.Fatal("Unexpected error converting back to client.Object after deep copy")
	}

	return copyObj
}
