package test

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

// testKubeClient implements a kubernetes.Client that uses
// a fake client.Client under the hood. It reimplements server-side
// apply since Fake client doesn't support it.
type testKubeClient struct {
	kubernetes.Client
	fakeClient client.Client
}

// ApplyServerSide creates or patches an object using server side apply logic.
// Uses client.Apply patch type like production code. Requires GVK to be set on object.
func (t *testKubeClient) ApplyServerSide(ctx context.Context, fieldManager string, obj kubernetes.Object, opts ...kubernetes.ApplyServerSideOption) error {
	o := &kubernetes.ApplyServerSideOptions{}
	for _, opt := range opts {
		opt.ApplyToApplyServerSide(o)
	}

	// Ensure GVK is set - client.Apply requires it
	if obj.GetObjectKind().GroupVersionKind().Empty() {
		gvks, _, err := t.fakeClient.Scheme().ObjectKinds(obj)
		if err != nil {
			return err
		}
		if len(gvks) > 0 {
			obj.GetObjectKind().SetGroupVersionKind(gvks[0])
		}
	}

	// Use client.Apply with proper options
	patchOpts := []client.PatchOption{
		client.FieldOwner(fieldManager),
	}
	if o.ForceOwnership {
		patchOpts = append(patchOpts, client.ForceOwnership)
	}

	return t.fakeClient.Patch(ctx, obj, client.Apply, patchOpts...)
}

// NewKubeClient builds a new kubernetes.Client by using client.Client.
func NewKubeClient(client client.Client) kubernetes.Client {
	return &testKubeClient{
		fakeClient: client,
		Client:     clientutil.NewKubeClient(client),
	}
}

// NewFakeKubeClient returns a kubernetes.Client that uses a fake client.Client under the hood.
func NewFakeKubeClient(objs ...client.Object) kubernetes.Client {
	return NewKubeClient(fake.NewClientBuilder().WithObjects(objs...).Build())
}

// NewFakeKubeClientAlwaysError returns a kubernetes.Client  that will always fail in any operation
// This is achieved by injecting an empty Scheme, which will make the underlying client.Client
// incapable of determining the resource type for a particular client.Object.
func NewFakeKubeClientAlwaysError(objs ...client.Object) kubernetes.Client {
	return NewKubeClient(
		fake.NewClientBuilder().WithScheme(runtime.NewScheme()).WithObjects(objs...).Build(),
	)
}
