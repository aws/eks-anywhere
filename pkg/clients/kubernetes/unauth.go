package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type KubectlGetter interface {
	GetObject(ctx context.Context, resourceType, name, namespace, kubeconfig string, obj runtime.Object) error
	Delete(ctx context.Context, resourceType, name, namespace, kubeconfig string) error
	Apply(ctx context.Context, kubeconfig string, obj runtime.Object) error
}

// UnAuthClient is a generic kubernetes API client that takes a kubeconfig
// file on every call in order to authenticate.
type UnAuthClient struct {
	kubectl KubectlGetter
	scheme  *runtime.Scheme
}

func NewUnAuthClient(kubectl KubectlGetter) *UnAuthClient {
	return &UnAuthClient{
		kubectl: kubectl,
		scheme:  runtime.NewScheme(),
	}
}

// Init initializes the client internal API scheme
// It has always be invoked at least once before making any API call
// It is not thread safe.
func (c *UnAuthClient) Init() error {
	return addToScheme(c.scheme, schemeAdders...)
}

// Get performs a GET call to the kube API server authenticating with a kubeconfig file
// and unmarshalls the response into the provdied Object.
func (c *UnAuthClient) Get(ctx context.Context, name, namespace, kubeconfig string, obj runtime.Object) error {
	resourceType, err := c.resourceTypeForObj(obj)
	if err != nil {
		return fmt.Errorf("getting kubernetes resource: %v", err)
	}

	return c.kubectl.GetObject(ctx, resourceType, name, namespace, kubeconfig, obj)
}

// KubeconfigClient returns an equivalent authenticated client.
func (c *UnAuthClient) KubeconfigClient(kubeconfig string) Client {
	return NewKubeconfigClient(c, kubeconfig)
}

// Delete performs a DELETE call to the kube API server authenticating with a kubeconfig file.
func (c *UnAuthClient) Delete(ctx context.Context, name, namespace, kubeconfig string, obj runtime.Object) error {
	resourceType, err := c.resourceTypeForObj(obj)
	if err != nil {
		return fmt.Errorf("deleting kubernetes resource: %v", err)
	}

	return c.kubectl.Delete(ctx, resourceType, name, namespace, kubeconfig)
}

func (c *UnAuthClient) Apply(ctx context.Context, kubeconfig string, obj runtime.Object) error {
	return c.kubectl.Apply(ctx, kubeconfig, obj)
}

func (c *UnAuthClient) resourceTypeForObj(obj runtime.Object) (string, error) {
	groupVersionKind, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return "", err
	}

	return groupVersionToKubectlResourceType(groupVersionKind), nil
}

func groupVersionToKubectlResourceType(g schema.GroupVersionKind) string {
	return fmt.Sprintf("%s.%s.%s", g.Kind, g.Version, g.Group)
}
