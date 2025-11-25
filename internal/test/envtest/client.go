package envtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateObjs creates Objects using the provided kube client and waits until its cache
// has been updated with those objects.
func CreateObjs(ctx context.Context, t testing.TB, c client.Client, objs ...client.Object) {
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

	newStatuses := []updatedStatus{}
	noStatusObjs := []client.Object{}

	for _, o := range objs {
		newStatus := updateStatus(ctx, t, c, o)
		if newStatus != nil {
			newStatuses = append(newStatuses, updatedStatus{
				obj:       o,
				newStatus: newStatus,
			})
		} else {
			noStatusObjs = append(noStatusObjs, o)
		}
	}

	// If the status doesn't need to be updated, just wait for the object to
	// to be available.
	for _, o := range noStatusObjs {
		waitForObjectAvailable(ctx, t, c, o)
	}

	for _, u := range newStatuses {
		waitForStatusUpdated(ctx, t, c, u.obj, u.newStatus)
	}
}

type updatedStatus struct {
	obj       client.Object
	newStatus map[string]interface{}
}

// UpdateStatusAndWait updates an objects status subresource and waits until the cache refreshes
// and reflects the new status.
func UpdateStatusAndWait(ctx context.Context, t testing.TB, c client.Client, o client.Object) {
	newStatus := updateStatus(ctx, t, c, o)
	if newStatus != nil {
		waitForStatusUpdated(ctx, t, c, o, newStatus)
	}
}

func updateStatus(ctx context.Context, t testing.TB, c client.Client, o client.Object) (newStatus map[string]interface{}) {
	objUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o)
	if err != nil {
		t.Fatalf("Failed converting object %s to unstructured: %v", klog.KObj(o), err)
	}
	obj := &unstructured.Unstructured{Object: objUnstructured}

	newStatus, found, err := unstructured.NestedMap(objUnstructured, "status")
	if err != nil {
		t.Fatalf("Failed checking status for object %s: %v", klog.KObj(obj), err)
	}
	if !found || len(newStatus) == 0 {
		return nil
	}

	objReady := waitForObjectAvailable(ctx, t, c, obj)

	// We need to update the status independently, kubernetes doesn't allow to create the main objects and
	// its subresources all at once
	obj.SetResourceVersion(objReady.GetResourceVersion())
	if err := c.Status().Update(ctx, obj); apierrors.IsNotFound(err) {
		// Some objects without a subresource will fail here,
		// so we just try and if it fails with a 404, we ignore the error
		t.Logf(
			"Try updating status but failed with a 404 error for [%s name=%s namespace=%s] object, most probably because it doesn't have a defined status subresource",
			obj.GetObjectKind().GroupVersionKind().String(),
			obj.GetName(),
			obj.GetNamespace(),
		)
	} else if err != nil {
		t.Fatal(err)
	}

	return newStatus
}

func waitForStatusUpdated(ctx context.Context, t testing.TB, c client.Client, o client.Object, newStatus map[string]interface{}) {
	g := gomega.NewWithT(t)
	g.Eventually(func(g gomega.Gomega) error {
		updatedObj := &unstructured.Unstructured{}
		updatedObj.SetGroupVersionKind(o.GetObjectKind().GroupVersionKind())
		g.Expect(
			c.Get(ctx, types.NamespacedName{Name: o.GetName(), Namespace: o.GetNamespace()}, updatedObj),
		).To(gomega.Succeed())

		updatedStatus, found, err := unstructured.NestedMap(updatedObj.Object, "status")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		if !found {
			return errors.New("no status found in updated object")
		}
		g.Expect(updatedStatus).To(gomega.Equal(newStatus), "updated status should be equal to desired status")

		return nil
	}, 5*time.Second).Should(gomega.Succeed(), "the status should be updated")
}

func waitForObjectAvailable(ctx context.Context, t testing.TB, c client.Client, obj client.Object) *unstructured.Unstructured {
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

func isNamespace(obj client.Object) bool {
	_, isNamespaceStruct := obj.(*corev1.Namespace)
	return isNamespaceStruct ||
		obj.GetObjectKind().GroupVersionKind().GroupKind() == corev1.SchemeGroupVersion.WithKind("Namespace").GroupKind()
}

func copyObject(t testing.TB, obj client.Object) client.Object {
	copyRuntimeObj := obj.DeepCopyObject()
	copyObj, ok := copyRuntimeObj.(client.Object)
	if !ok {
		t.Fatal("Unexpected error converting back to client.Object after deep copy")
	}

	return copyObj
}

// APIExpecter is a helper to define eventual expectations over API resources in tests.
// It's useful when working with clients that maintain a cache, since changes might not be
// reflected immediately, causing tests to flake.
type APIExpecter struct {
	t       testing.TB
	client  client.Client
	g       gomega.Gomega
	timeout time.Duration
}

// NewAPIExpecter constructs a new APIExpecter.
func NewAPIExpecter(t testing.TB, client client.Client) *APIExpecter {
	return &APIExpecter{
		t:       t,
		g:       gomega.NewWithT(t),
		client:  client,
		timeout: 5 * time.Second,
	}
}

// DeleteAndWait sends delete requests for a collection of objects and waits until
// the client cache reflects the changes.
func (a *APIExpecter) DeleteAndWait(ctx context.Context, objs ...client.Object) {
	a.t.Helper()
	for _, obj := range objs {
		// namespaces can't be deleted with envtest
		if isNamespace(obj) {
			continue
		}

		err := a.client.Delete(ctx, obj)
		if !apierrors.IsNotFound(err) {
			a.g.Expect(err).To(gomega.Succeed(), "should delete object %s", obj.GetName())
		}
		a.ShouldEventuallyNotExist(ctx, obj)
	}
}

// DeleteAllOfAndWait deletes all objects of the given type and waits until the client's
// cache reflects those changes.
func (a *APIExpecter) DeleteAllOfAndWait(ctx context.Context, obj client.Object) {
	a.t.Helper()
	a.g.Eventually(func() error {
		err := a.client.DeleteAllOf(ctx, obj)
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		return errors.New("some objects still existed before delete operation, try deleting another round")
	}, a.timeout).Should(gomega.Succeed(), "all objects of kind %s should eventually be deleted", obj.GetObjectKind().GroupVersionKind().Kind)
}

// ShouldEventuallyExist defines an eventual expectation that succeeds if the provided object
// becomes readable by the client before the timeout expires.
func (a *APIExpecter) ShouldEventuallyExist(ctx context.Context, obj client.Object) {
	a.t.Helper()
	key := client.ObjectKeyFromObject(obj)
	a.g.Eventually(func() error {
		return a.client.Get(ctx, key, obj)
	}, a.timeout).Should(gomega.Succeed(), "object %s should eventually exist", obj.GetName())
}

// ShouldEventuallyMatch defines an eventual expectation that succeeds if the provided object
// becomes readable by the client and matches the provider expectation before the timeout expires.
func (a *APIExpecter) ShouldEventuallyMatch(ctx context.Context, obj client.Object, match func(g gomega.Gomega)) {
	a.t.Helper()
	key := client.ObjectKeyFromObject(obj)
	a.g.Eventually(func(g gomega.Gomega) error {
		if err := a.client.Get(ctx, key, obj); err != nil {
			return err
		}

		match(g)

		return nil
	}, a.timeout).Should(gomega.Succeed(), "object %s should eventually match", obj.GetName())
}

// CloneNameNamespace returns an empty client object of the same type
// with the same and namespace. This is a helper to pass a new object to the "Eventually"
// methods while preserving the original object's data.
func CloneNameNamespace[T any, PT interface {
	*T
	client.Object
}](obj PT,
) PT {
	copyObj := PT(new(T))
	copyObj.SetName(obj.GetName())
	copyObj.SetNamespace(obj.GetNamespace())
	return copyObj
}

// ShouldEventuallyNotExist defines an eventual expectation that succeeds if the provided object
// becomes not found by the client before the timeout expires.
func (a *APIExpecter) ShouldEventuallyNotExist(ctx context.Context, obj client.Object) {
	key := client.ObjectKeyFromObject(obj)
	a.g.Eventually(func() error {
		err := a.client.Get(ctx, key, obj)
		if apierrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		return errors.New("object still exists")
	}, a.timeout).Should(gomega.Succeed(), "object %s should eventually be deleted", obj.GetName())
}
