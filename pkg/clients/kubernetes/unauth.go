package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

// UnAuthClient is a generic kubernetes API client that takes a kubeconfig
// file on every call in order to authenticate.
type UnAuthClient struct {
	kubectl Kubectl
	scheme  *Scheme
}

// NewUnAuthClient builds a new UnAuthClient.
func NewUnAuthClient(kubectl Kubectl) *UnAuthClient {
	return &UnAuthClient{
		kubectl: kubectl,
		scheme:  &Scheme{runtime.NewScheme()},
	}
}

// Init initializes the client internal API scheme
// It has always be invoked at least once before making any API call
// It is not thread safe.
func (c *UnAuthClient) Init() error {
	return InitScheme(c.scheme.Scheme)
}

// Get performs a GET call to the kube API server authenticating with a kubeconfig file
// and unmarshalls the response into the provdied Object.
func (c *UnAuthClient) Get(ctx context.Context, name, namespace, kubeconfig string, obj runtime.Object) error {
	resourceType, err := c.resourceTypeForObj(obj)
	if err != nil {
		return fmt.Errorf("getting kubernetes resource: %v", err)
	}

	var opts KubectlGetOptions

	if namespace == "" {
		opts = KubectlGetOptions{Name: name, ClusterScoped: pointer.Bool(true)}
	} else {
		opts = KubectlGetOptions{Name: name, Namespace: namespace}
	}

	return c.kubectl.Get(ctx, resourceType, kubeconfig, obj, &opts)
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
func (c *UnAuthClient) List(ctx context.Context, kubeconfig string, list ObjectList, opts ...ListOption) error {
	if len(opts) > 0 {
		return fmt.Errorf("list options are not supported for unauthenticated clients")
	}
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
	return c.scheme.KubectlResourceTypeForObj(obj)
}
