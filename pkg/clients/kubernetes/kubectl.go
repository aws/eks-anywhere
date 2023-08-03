package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

// Kubectl is a client implemented with the kubectl binary.
type Kubectl interface {
	Get(ctx context.Context, resourceType, kubeconfig string, obj runtime.Object, opts ...KubectlGetOption) error
	Create(ctx context.Context, kubeconfig string, obj runtime.Object) error
	Replace(ctx context.Context, kubeconfig string, obj runtime.Object) error
	Apply(ctx context.Context, kubeconfig string, obj runtime.Object, opts ...KubectlApplyOption) error
	Delete(ctx context.Context, resourceType, kubeconfig string, opts ...KubectlDeleteOption) error
}

// KubectlGetOption is some configuration that modifies options for a get request.
type KubectlGetOption interface {
	// ApplyToGet applies this configuration to the given get options.
	ApplyToGet(*KubectlGetOptions)
}

// KubectlGetOptions contains options for get commands.
type KubectlGetOptions struct {
	// Name specifies the name of a resource. If set, only one single resource
	// will be returned (at most). If set, Namespace is required.
	Name string

	// Namespace specifies the namespace to retrieve objects from. If not set,
	// all namespaces will be used.
	Namespace string

	// ClusterScoped identifies the resourced as no namespaced. This is mutually exclusive with
	// Namespace and requires to also specify a Name.
	ClusterScoped *bool
}

var _ KubectlGetOption = &KubectlGetOptions{}

// ApplyToGet applies this configuration to the given get options.
func (o *KubectlGetOptions) ApplyToGet(kgo *KubectlGetOptions) {
	if o.Name != "" {
		kgo.Name = o.Name
	}
	if o.Namespace != "" {
		kgo.Namespace = o.Namespace
	}
	if o.ClusterScoped != nil {
		kgo.ClusterScoped = o.ClusterScoped
	}
}

// KubectlApplyOption is some configuration that modifies options for an apply command.
type KubectlApplyOption interface {
	// ApplyToApply applies this configuration to the given apply options.
	ApplyToApply(*KubectlApplyOptions)
}

// KubectlApplyOptions contains options for apply command.
type KubectlApplyOptions struct {
	// ServerSide configures kubectl so apply runs in the server instead of the client.
	ServerSide bool

	// ForceOwnership indicates that in case of conflicts with server-side apply,
	// kubectl should acquire ownership of the conflicting field.
	ForceOwnership bool

	// FieldManager sets the field owner name for the given server-side apply.
	FieldManager string
}

var _ KubectlApplyOption = KubectlApplyOptions{}

// ApplyToApply applies this configuration to the given apply options.
func (o KubectlApplyOptions) ApplyToApply(kao *KubectlApplyOptions) {
	if o.ServerSide {
		kao.ServerSide = true
	}

	if o.ForceOwnership {
		kao.ForceOwnership = true
	}

	if o.FieldManager != "" {
		kao.FieldManager = o.FieldManager
	}
}

// KubectlDeleteOption is some configuration that modifies options for a delete command.
type KubectlDeleteOption interface {
	// ApplyToDelete applies this configuration to the given delete options.
	ApplyToDelete(*KubectlDeleteOptions)
}

// KubectlDeleteOptions contains options for delete commands.
type KubectlDeleteOptions struct {
	// Name specifies the name of a resource. Use to delete a single resource.
	// If set, Namespace is required.
	Name string

	// Namespace specifies the namespace to delete objects from. If not set,
	// all namespaces will be used.
	Namespace string

	// HasLabels applies a filter using labels to the objects to be deleted.
	// When multiple label-value pairs are specified, the condition is an AND
	// for all of them. If specified, Name should be empty.
	HasLabels map[string]string
}

var _ KubectlDeleteOption = &KubectlDeleteOptions{}

// ApplyToDelete applies this configuration to the given delete options.
func (o *KubectlDeleteOptions) ApplyToDelete(kdo *KubectlDeleteOptions) {
	if o.Name != "" {
		kdo.Name = o.Name
	}
	if o.Namespace != "" {
		kdo.Namespace = o.Namespace
	}
	if o.HasLabels != nil {
		kdo.HasLabels = o.HasLabels
	}
}
