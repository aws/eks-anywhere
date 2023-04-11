package test

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

// NewKubeClient builds a new kubernetes.Client by using client.Client.
func NewKubeClient(client client.Client) kubernetes.Client {
	return clientutil.NewKubeClient(client)
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
