package test

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

// ApplyServerSide creates or patches and object using server side logic.
// Giving the limitations of the fake client, we implement a fake server side apply
// using a simplified version, using a raw update if the object exists and create
// otherwise.
func (t *testKubeClient) ApplyServerSide(ctx context.Context, fieldManager string, obj kubernetes.Object, opts ...kubernetes.ApplyServerSideOption) error {
	err := t.fakeClient.Update(ctx, obj)
	if apierrors.IsNotFound(err) {
		return t.fakeClient.Create(ctx, obj)
	}

	return err
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
