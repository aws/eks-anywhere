package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Object is a Kubernetes object.
type Object client.Object

// ObjectList is a Kubernetes object list.
type ObjectList client.ObjectList

// Client is Kubernetes API client.
type Client interface {
	Reader
	Writer
}

// Reader knows how to read and list Kubernetes objects.
type Reader interface {
	// Get retrieves an obj for the given name and namespace from the Kubernetes Cluster.
	Get(ctx context.Context, name, namespace string, obj Object) error

	// List retrieves list of objects. On a successful call, Items field
	// in the list will be populated with the result returned from the server.
	List(ctx context.Context, list ObjectList, opt ...ListOption) error
}

// Writer knows how to create, delete, and update Kubernetes objects.
type Writer interface {
	// Create saves the object obj in the Kubernetes cluster.
	Create(ctx context.Context, obj Object) error

	// Update updates the given obj in the Kubernetes cluster.
	Update(ctx context.Context, obj Object) error

	// ApplyServerSide creates or patches and object using server side logic.
	ApplyServerSide(ctx context.Context, fieldManager string, obj Object, opts ...ApplyServerSideOption) error

	// Delete deletes the given obj from Kubernetes cluster.
	Delete(ctx context.Context, obj Object) error

	// DeleteAllOf deletes all objects of the given type matching the given options.
	DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error
}

// DeleteAllOfOption is some configuration that modifies options for a delete request.
type DeleteAllOfOption interface {
	// ApplyToDeleteAllOf applies this configuration to the given deletecollection options.
	ApplyToDeleteAllOf(*DeleteAllOfOptions)
}

// DeleteAllOfOptions contains options for deletecollection (deleteallof) requests.
type DeleteAllOfOptions struct {
	// HasLabels filters results by label and value. The requirement is an AND match
	// for all labels.
	HasLabels map[string]string

	// Namespace represents the namespace to list for, or empty for
	// non-namespaced objects, or to list across all namespaces.
	Namespace string
}

var _ DeleteAllOfOption = &DeleteAllOfOptions{}

// ApplyToDeleteAllOf implements DeleteAllOfOption.
func (o *DeleteAllOfOptions) ApplyToDeleteAllOf(do *DeleteAllOfOptions) {
	if o.HasLabels != nil {
		do.HasLabels = o.HasLabels
	}
	if o.Namespace != "" {
		do.Namespace = o.Namespace
	}
}

// ApplyServerSideOption is some configuration that modifies options for an apply request.
type ApplyServerSideOption interface {
	ApplyToApplyServerSide(*ApplyServerSideOptions)
}

// ApplyServerSideOptions contains options for server side apply requests.
type ApplyServerSideOptions struct {
	// ForceOwnership indicates that in case of conflicts with server-side apply,
	// the client should acquire ownership of the conflicting field.
	ForceOwnership bool
}

var _ ApplyServerSideOption = ApplyServerSideOptions{}

// ApplyToApplyServerSide implements ApplyServerSideOption.
func (o ApplyServerSideOptions) ApplyToApplyServerSide(do *ApplyServerSideOptions) {
	if o.ForceOwnership {
		do.ForceOwnership = true
	}
}

// ListOptions contains options for list requests.
type ListOptions struct {
	// LabelSelector filters results by label. Use labels.Parse() to
	// set from raw string form.
	LabelSelector labels.Selector
}

// ListOption is some configuration that modifies options for a List request.
type ListOption interface {
	ApplyToList(*ListOptions)
}

// ApplyToList implements ListOption.
func (o ListOptions) ApplyToList(lo *ListOptions) {
	if o.LabelSelector != nil {
		lo.LabelSelector = o.LabelSelector
	}
}
