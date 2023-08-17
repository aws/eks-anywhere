package kubernetes

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// UnAuthClient is a generic kubernetes API client that takes a kubeconfig
// file on every call in order to authenticate.
type UnAuthClient struct {
	kubectl Kubectl
	scheme  *runtime.Scheme
}

// NewUnAuthClient builds a new UnAuthClient.
func NewUnAuthClient(kubectl Kubectl) *UnAuthClient {
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

	return c.kubectl.Get(ctx, resourceType, kubeconfig, obj, &KubectlGetOptions{Name: name, Namespace: namespace})
}

// KubeconfigClient returns an equivalent authenticated client.
func (c *UnAuthClient) KubeconfigClient(kubeconfig string) Client {
	return NewKubeconfigClient(c, kubeconfig)
}

// BuildClientFromKubeconfig returns an equivalent authenticated client. It will never return
// an error but this helps satisfy a generic factory interface where errors are possible. It's
// basically an alias to KubeconfigClient.
func (c *UnAuthClient) BuildClientFromKubeconfig(kubeconfig string) (Client, error) {
	return c.KubeconfigClient(kubeconfig), nil
}

// Apply performs an upsert in the form of a client-side apply.
func (c *UnAuthClient) Apply(ctx context.Context, kubeconfig string, obj runtime.Object) error {
	return c.kubectl.Apply(ctx, kubeconfig, obj)
}

// ApplyServerSide creates or patches and object using server side logic.
func (c *UnAuthClient) ApplyServerSide(ctx context.Context, kubeconfig, fieldManager string, obj Object, opts ...ApplyServerSideOption) error {
	o := &ApplyServerSideOptions{}
	for _, opt := range opts {
		opt.ApplyToApplyServerSide(o)
	}

	ko := KubectlApplyOptions{
		ServerSide:   true,
		FieldManager: fieldManager,
	}
	if o.ForceOwnership {
		ko.ForceOwnership = o.ForceOwnership
	}

	return c.kubectl.Apply(ctx, kubeconfig, obj, ko)
}

// List retrieves list of objects. On a successful call, Items field
// in the list will be populated with the result returned from the server.
func (c *UnAuthClient) List(ctx context.Context, kubeconfig string, list ObjectList) error {
	resourceType, err := c.resourceTypeForObj(list)
	if err != nil {
		return fmt.Errorf("getting kubernetes resource: %v", err)
	}

	return c.kubectl.Get(ctx, resourceType, kubeconfig, list)
}

// Create saves the object obj in the Kubernetes cluster.
func (c *UnAuthClient) Create(ctx context.Context, kubeconfig string, obj Object) error {
	return c.kubectl.Create(ctx, kubeconfig, obj)
}

// Update updates the given obj in the Kubernetes cluster.
func (c *UnAuthClient) Update(ctx context.Context, kubeconfig string, obj Object) error {
	return c.kubectl.Replace(ctx, kubeconfig, obj)
}

// Delete deletes the given obj from Kubernetes cluster.
func (c *UnAuthClient) Delete(ctx context.Context, kubeconfig string, obj Object) error {
	resourceType, err := c.resourceTypeForObj(obj)
	if err != nil {
		return fmt.Errorf("deleting kubernetes resource: %v", err)
	}

	o := &KubectlDeleteOptions{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	return c.kubectl.Delete(ctx, resourceType, kubeconfig, o)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (c *UnAuthClient) DeleteAllOf(ctx context.Context, kubeconfig string, obj Object, opts ...DeleteAllOfOption) error {
	resourceType, err := c.resourceTypeForObj(obj)
	if err != nil {
		return fmt.Errorf("deleting kubernetes resource: %v", err)
	}

	deleteAllOpts := &DeleteAllOfOptions{}
	for _, opt := range opts {
		opt.ApplyToDeleteAllOf(deleteAllOpts)
	}

	o := &KubectlDeleteOptions{}
	o.Namespace = deleteAllOpts.Namespace
	o.HasLabels = deleteAllOpts.HasLabels
	return c.kubectl.Delete(ctx, resourceType, kubeconfig, o)
}

func (c *UnAuthClient) resourceTypeForObj(obj runtime.Object) (string, error) {
	groupVersionKind, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return "", err
	}

	if meta.IsListType(obj) && strings.HasSuffix(groupVersionKind.Kind, "List") {
		// if obj is a list, treat it as a request for the "individual" item's resource
		groupVersionKind.Kind = groupVersionKind.Kind[:len(groupVersionKind.Kind)-4]
	}

	return groupVersionToKubectlResourceType(groupVersionKind), nil
}

func groupVersionToKubectlResourceType(g schema.GroupVersionKind) string {
	if g.Group == "" {
		// if Group is not set, this probably an obj from "core", which api group is just v1
		return g.Kind
	}

	return fmt.Sprintf("%s.%s.%s", g.Kind, g.Version, g.Group)
}
